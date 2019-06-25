package barrelman

import (
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

type DescribeCmd struct {
	Options         *CmdOptions
	Config          *Config
	ManifestName    string
	ManifestVersion int32
	LogOptions      *[]string
}

type DescribeTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	Chart       *cluster.Chart
	State       int
	Diff        []byte
	Changed     bool
}

type DescribeTargets struct {
	ManifestName string
	Data         []*ReleaseTarget
}

func (cmd *DescribeCmd) Run(session cluster.Sessioner) error {
	var err error
	log.Rep(version.Get()).Info("Barrelman")

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

	versionTable := versions.Table()

	if cmd.ManifestVersion != 0 {
		version, ok := versionTable.Data[cmd.ManifestVersion]
		if !ok {
			return errors.WithFields(errors.Fields{
				"ManifestVersion": cmd.ManifestVersion,
				"ManifestName":    cmd.ManifestName,
			}).New("Failed to rollback to version, No such version")
		}
		releaseTable, err := version.ReleaseTable()
		if err != nil {
			return errors.Wrap(err, "failed to get release table from Rollback ConfigMap")
		}
		for k, v := range releaseTable {
			log.WithFields(log.Fields{
				"ManifestVersion": cmd.ManifestVersion,
				"ManifestName":    cmd.ManifestName,
				"ReleaseName":     k,
				"Revision":        v.Value,
			}).Info("Rollback release")
		}
	}
	return nil
}
