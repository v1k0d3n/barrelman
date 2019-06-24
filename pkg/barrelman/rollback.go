package barrelman

import (
	"fmt"
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
	ReleaseMeta     *cluster.ReleaseMeta
	TransitionState TransitionState
	ReleaseVersion  *cluster.Version
	Revision        int32
	Diff            []byte
	Changed         bool
}

type RollbackTargets struct {
	ManifestName string
	session      cluster.Sessioner
	transaction  cluster.Transactioner
	Data         []*RollbackTarget
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
		}).Debug("Rollback manifest")
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

		currentReleases, err := session.ReleasesByManifest(cmd.ManifestName)
		if err != nil {
			return errors.Wrap(err, "failed to get current releases")
		}
		rts, err := cmd.ComputeRollback(session, transaction, releaseTable, currentReleases)
		if err != nil {
			return errors.Wrap(err, "failed to compute plan for rollback")
		}

		_, err = rts.Diff(session)
		if err != nil {
			return err
		}

		if cmd.Options.Diff {
			rts.LogDiff()
			return nil
		}

		if err := rts.Apply(); err != nil {
			return errors.Wrap(err, "Rollback failed")
		}
	}

	return transaction.Complete()
}

func (rts *RollbackTargets) Apply() error {

	for _, rt := range rts.Data {
		switch rt.TransitionState {
		case NoChange:
			log.WithFields(log.Fields{
				"ReleaseName": rt.ReleaseMeta.ReleaseName,
			}).Debug("No change in rollback")
		case Installable:
			return errors.WithFields(errors.Fields{
				"ReleaseName":     rt.ReleaseMeta.ReleaseName,
				"TransitionState": rt.TransitionState.String(),
			}).New("Invalid transition state for rollback")
		case Upgradable, Undeletable, Replaceable:
			newRevision, err := rts.session.RollbackRelease(&cluster.RollbackMeta{
				ReleaseName: rt.ReleaseMeta.ReleaseName,
				Revision:    rt.Revision,
			})
			if err != nil {
				return errors.Wrap(err, "rollback failed")
			}
			log.WithFields(log.Fields{
				"ReleaseName": rt.ReleaseMeta.ReleaseName,
				"Version":     newRevision,
			}).Info("Release rolled back")
			rt.ReleaseVersion.SetModified()
		case Deletable:
			log.WithFields(log.Fields{
				"ReleaseName": rt.ReleaseMeta.ReleaseName,
			}).Info("Rollback to deleted")
			if err := rts.session.DeleteRelease(&cluster.DeleteMeta{
				ReleaseName: rt.ReleaseMeta.ReleaseName,
				Namespace:   rt.ReleaseVersion.Namespace,
			}); err != nil {
				return errors.Wrap(err, "error deleting release during rollback")
			}
			rt.ReleaseVersion.SetModified()
		default:
			// Not a thing
			return errors.WithFields(errors.Fields{
				"ReleaseName":     rt.ReleaseMeta.ReleaseName,
				"TransitionState": rt.TransitionState.String(),
			}).New("Unhandled transition state")
		}
	}
	return nil
}

//ComputeReleases configures each potential release with a current state
//states may be one of 'Installable', 'Upgradeable', 'Replaceable', 'NoChange'
func (cmd *RollbackCmd) ComputeRollback(
	session cluster.Sessioner,
	transaction cluster.Transactioner,
	rollbackReleaseList map[string]*chart.Value,
	currentReleases map[string]*cluster.ReleaseMeta) (*RollbackTargets, error) {

	rts := &RollbackTargets{
		ManifestName: cmd.ManifestName,
		session:      session,
		transaction:  transaction,
	}

	for releaseName, v := range rollbackReleaseList {
		releaseExists := false
		tmpVersion, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to extract release version from rollback data")
		}
		revision := int32(tmpVersion)

		rt := &RollbackTarget{
			ReleaseMeta: &cluster.ReleaseMeta{
				ReleaseName: releaseName,
			},
			Revision:        revision,
			TransitionState: NoChange, //Unless modified below
		}

		//Evaluate rollback vs current releases
		for _, rel := range currentReleases {
			if rel.ReleaseName == rt.ReleaseMeta.ReleaseName {
				releaseExists = true

				rt.ReleaseVersion = &cluster.Version{
					Name:      rel.ReleaseName,
					Namespace: rel.Namespace,
					Revision:  rel.Revision,
				}

				// The To Chart is needed to perform diffs
				// not technically needed for the rollback
				// but for operator analysis and avoiding rolling on no change
				toMeta, err := session.GetRelease(rel.ReleaseName, revision)
				if err != nil {
					return nil, errors.Wrap(err, "failed to get release")
				}
				rt.ReleaseVersion.Chart = toMeta.Chart
				rt.ReleaseMeta.Config = toMeta.Config
				if err := rt.CalculateDiff(session); err != nil {
					return nil, err
				}
				if rel.Status == cluster.Status_DELETED {
					// Current release has been deleted, we track it seperately
					rt.TransitionState = Undeletable
				} else if rel.Status == cluster.Status_FAILED {
					// Current release is in FAILED state AND force is enabled for this release
					// setup for delete and install
					rt.TransitionState = Replaceable
				} else {
					// Set to upgradeable, then check for difference
					rt.TransitionState = Upgradable
					if err := rt.CalculateDiff(session); err != nil {
						return nil, err
					}
					if !rt.Changed {
						rt.TransitionState = NoChange
					}
				}
			}
		}
		if !releaseExists {
			//There is no existing releases, we can't roll to this
			return nil, errors.WithFields(errors.Fields{
				"ReleaseName": rt.ReleaseMeta.ReleaseName,
			}).New("Cannot roll to release, it doesn't exist")
		}

		log.WithFields(log.Fields{
			"ReleaseName":     rt.ReleaseMeta.ReleaseName,
			"TransitionState": rt.TransitionState.String(),
		}).Debug("computed transition state")
		rts.Data = append(rts.Data, rt)

		// Add this release tartget to the transaction
		transaction.Versions().AddReleaseVersion(rt.ReleaseVersion)
	}
	// iterate current releases, any that do not exist in rollbackReleaseList set to delete
	for _, rel := range currentReleases {
		if rel.Status == cluster.Status_DELETED {
			log.WithFields(log.Fields{
				"RunningReleaseName": rel.ReleaseName,
			}).Debug("Already deleted")
			// Already deleted, so noop
			continue
		}

		// if the current release exists in rollback, move on
		if _, ok := rollbackReleaseList[rel.ReleaseName]; ok {
			continue
		}

		rv := &cluster.Version{
			Name:      rel.ReleaseName,
			Namespace: rel.Namespace,
		}
		rts.Data = append(rts.Data, &RollbackTarget{
			ReleaseMeta: &cluster.ReleaseMeta{
				ReleaseName: rel.ReleaseName,
			},
			ReleaseVersion:  rv,
			TransitionState: Deletable,
		})
		rv.SetModified()
	}
	return rts, nil
}

func (rts *RollbackTargets) Diff(session cluster.Sessioner) (*RollbackTargets, error) {
	for _, v := range rts.Data {
		if err := v.CalculateDiff(session); err != nil {
			return rts, err
		}
	}
	return rts, nil
}

func (rt *RollbackTarget) CalculateDiff(session cluster.Sessioner) error {
	var err error

	rt.ReleaseMeta.DryRun = true
	switch rt.TransitionState {
	case Upgradable, Replaceable:
		rt.Changed, rt.Diff, err = session.DiffRelease(&cluster.ReleaseMeta{
			Chart:          rt.ReleaseVersion.Chart,
			ReleaseName:    rt.ReleaseVersion.Name,
			Namespace:      rt.ReleaseVersion.Namespace,
			ValueOverrides: []byte(rt.ReleaseMeta.Config.Raw),
		})
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"Changed": rt.Changed,
			"Name":    rt.ReleaseVersion.Name,
		}).Warn("diff verdict")
	}
	return nil
}

func (rt *RollbackTargets) LogDiff() {
	for _, v := range rt.Data {
		switch v.TransitionState {
		case Deletable:
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Would delete")
		case Installable, Undeletable:
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Would install")
		case Upgradable:
			if v.Changed {
				log.WithFields(log.Fields{
					"Name": v.ReleaseMeta.ReleaseName,
				}).Info("Diff")
				//Print the byte content and keep formatting, its fancy
				fmt.Printf("----%v\n%v_____\n", v.ReleaseMeta.MetaName, string(v.Diff))
			} else {
				log.WithFields(log.Fields{
					"Name": v.ReleaseMeta.ReleaseName,
				}).Info("No change")
			}
		}
	}
}
