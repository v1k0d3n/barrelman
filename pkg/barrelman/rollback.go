package barrelman

import (
	"strconv"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type RollbackCmd struct {
	Options         *CmdOptions
	Config          *Config
	ManifestName    string
	ManifestVersion int32
	LogOptions      *[]string
}

type RollbackTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	Chart       *cluster.Chart
	State       int
	Diff        []byte
	Changed     bool
}

type RollbackTargets struct {
	ManifestName string
	Data         []*ReleaseTarget
}

func (cmd *RollbackCmd) Run(session cluster.Sessioner) error {
	var err error
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	log.Debug("connecting to cluster")
	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	if session.GetKubeConfig() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeConfig(),
		}).Info("Using kube config")
	}
	if session.GetKubeContext() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}

	versions, err := session.GetVersions(cmd.ManifestName)
	if err != nil {
		return errors.Wrap(err, "Failed to get versions")
	}
	for _, v := range versions.Data {
		log.WithFields(log.Fields{
			"manifestName": v.Name,
			"namespace":    v.Namespace,
			"Revision":     v.Revision,
		}).Info("Rollback manifest")
	}

	// Rollback supports transactions
	transaction, err := session.NewTransaction(cmd.ManifestName)
	if err != nil {
		return errors.Wrap(err, "failed to create new transaction durring apply")
	}
	defer transaction.Cancel()

	versionTable := versions.Table()
	if cmd.ManifestVersion != 0 {
		releaseMeta, ok := versionTable.Data[cmd.ManifestVersion]
		if !ok {
			return errors.WithFields(errors.Fields{
				"ManifestVersion": cmd.ManifestVersion,
				"ManifestName":    cmd.ManifestName,
			}).New("Failed to rollback to version, No such version")
		}
		releaseTable, err := releaseMeta.ReleaseTable()
		if err != nil {
			return errors.Wrap(err, "failed to get release table from Rollback ConfigMap")
		}
		for k, v := range releaseTable {
			log.WithFields(log.Fields{
				"ReleaseName": k,
				"Revision":    v.Value,
			}).Debug("Rollback release")
		}

		if err := cmd.deleteMissing(session, releaseMeta.Chart.Values.Values); err != nil {
			return errors.Wrap(err, "failed to remove release that does not exist in rollback")
		}

		for releaseName, releaseVersion := range releaseMeta.Chart.Values.Values {

			//Convert the *chart.Value to int32
			tmpVersion, err := strconv.Atoi(releaseVersion.Value)
			if err != nil {
				return errors.Wrap(err, "Failed to extract release version from rollback data")
			}
			releaseVersion := int32(tmpVersion)

			log.WithFields(log.Fields{
				"key":   releaseName,
				"value": releaseVersion,
			}).Debug("Rollback entry")
			newVersion, err := session.RollbackRelease(&cluster.RollbackMeta{
				ReleaseName: releaseName,
				Revision:    releaseVersion,
			})
			if err != nil {
				return errors.Wrap(err, "Rollback of release failed")
			}
			if err := transaction.Versions().AddReleaseVersion(&cluster.Version{
				Name:     releaseName,
				Revision: newVersion,
			}); err != nil {
				return errors.Wrap(err, "Failed to add release to transaction during rollback")
			}
			transaction.SetChanged()
		}
	}

	return transaction.Complete()
}

func (cmd *RollbackCmd) deleteMissing(session cluster.Sessioner, releaseList map[string]*chart.Value) error {
	list, err := session.ReleasesByManifest(cmd.ManifestName)
	if err != nil {
		return errors.Wrap(err, "Failed to get releases")
	}
	for k, v := range list {
		log.WithFields(log.Fields{
			"k":           k,
			"ReleaseName": v.ReleaseName,
		}).Debug("evaluating release for removal")
		if _, ok := releaseList[k]; !ok {
			if err := session.DeleteRelease(&cluster.DeleteMeta{
				ReleaseName: v.ReleaseName,
			}); err != nil {
				return err
			}
			log.WithFields(log.Fields{
				"ReleaseName": v.ReleaseName,
			}).Debug("deleted release")
		}
	}

	return nil
}
