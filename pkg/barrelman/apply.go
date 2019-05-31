package barrelman

import (
	"fmt"
	"time"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

type ApplyCmd struct {
	Options    *CmdOptions
	Config     *Config
	LogOptions *[]string
}

type ReleaseTarget struct {
	ReleaseMeta     *cluster.ReleaseMeta
	Chart           *cluster.Chart
	TransitionState TransitionState
	Diff            []byte
	Changed         bool
	ReleaseVersion  *cluster.Version
}

type ReleaseTargets struct {
	ManifestName string
	session      cluster.Sessioner
	transaction  cluster.Transactioner
	Data         []*ReleaseTarget
}

func (cmd *ApplyCmd) Run(session cluster.Sessioner) error {
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

	archives, manifestName, err := processManifest(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "apply failed")
	}

	transaction, err := session.NewTransaction(manifestName)
	if err != nil {
		return errors.Wrap(err, "failed to create new transaction durring apply")
	}

	releases, err := session.ReleasesByManifest(manifestName)
	if err != nil {
		return errors.Wrap(err, "failed to get current releases")
	}

	rt, err := cmd.ComputeReleases(session, transaction, manifestName, archives, releases)
	if err != nil {
		return err
	}

	if err := rt.dryRun(session); err != nil {
		return err
	}
	if cmd.Options.DryRun {
		log.Info("No errors")
		return nil
	}

	_, err = rt.Diff(session)
	if err != nil {
		return err
	}
	if cmd.Options.Diff {
		rt.LogDiff()
		return nil
	}

	for _, v := range rt.Data {
		log.WithFields(log.Fields{
			"ReleaseMetaRevision": v.ReleaseMeta.Revision,
			"ReleaseVersion":      v.ReleaseVersion.Revision,
			"ReleaseName":         v.ReleaseMeta.ReleaseName,
		}).Warn("Data")
	}

	err = rt.Apply(cmd.Options)
	if err != nil {
		if innerErr := transaction.Cancel(); innerErr != nil {
			err = errors.WithFields(errors.Fields{
				"TransactionError": innerErr.Error(),
			}).Wrap(err, "transaction error while Canceling")
		}
		return errors.Wrap(err, "Manifest upgrade failed")
	}

	return transaction.Complete()
}

//IsReplaceable checks a release against the --force flag values to see if an existing release should be replaced via delete
func (cmd *ApplyCmd) isInForce(rel *cluster.ReleaseMeta) bool {
	//Checks for releases configured for Force by cmdline
	for _, r := range *cmd.Options.Force {
		if r == rel.MetaName || r == rel.ChartName || r == rel.ReleaseName {
			return true
		}
	}
	return false
}

//ComputeReleases configures each potential release with a current state
//states may be one of 'Installable', 'Upgradeable', 'Replaceable', 'NoChange'
func (cmd *ApplyCmd) ComputeReleases(
	session cluster.Sessioner,
	transaction cluster.Transactioner,
	manifestName string,
	archives *manifest.ArchiveFiles,
	currentReleases map[string]*cluster.ReleaseMeta) (*ReleaseTargets, error) {
	rts := &ReleaseTargets{
		ManifestName: manifestName,
		session:      session,
		transaction:  transaction,
	}

	//These archives are created from the manifest, they may potentially be installed/upgraded
	for _, v := range archives.List {
		releaseExists := false
		inChart, err := session.ChartFromArchive(v.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate chart from archive")
		}

		rt := &ReleaseTarget{
			TransitionState: NoChange, //Unless modified below
			ReleaseMeta: &cluster.ReleaseMeta{
				Chart:          inChart,
				ReleaseName:    v.ReleaseName,
				Namespace:      v.Namespace,
				ValueOverrides: v.Overrides,
				InstallWait:    v.InstallWait,
			},
		}

		//Evaluate archive vs current releases
		for _, rel := range currentReleases {
			if rel.ReleaseName == v.ReleaseName {
				rt.ReleaseVersion = &cluster.Version{
					Name:             rel.ReleaseName,
					Namespace:        rel.Namespace,
					Revision:         rel.Revision,
					PreviousRevision: rel.Revision,
				}
				releaseExists = true
				if rel.Status == cluster.Status_DELETED {
					// Current release has been deleted, a state that is resisitant to Upgrade/Install
					// Rollback to the current revision, then Upgrade
					rt.TransitionState = Undeleteable
				} else if cmd.isInForce(rel) || rel.Status == cluster.Status_FAILED {
					// Current release is in FAILED state AND force is enabled for this release
					// setup for delete and install
					rt.TransitionState = Replaceable
				} else {
					// All other cases use Upgrade
					rt.TransitionState = Upgradable
				}
			}
		}
		if !releaseExists {
			//There is no existing releases, just Install
			rt.TransitionState = Installable
			rt.ReleaseVersion = &cluster.Version{
				Name:      v.ReleaseName,
				Namespace: v.Namespace,
			}
		}
		log.WithFields(log.Fields{
			"ReleaseName": v.ReleaseName,
			"Revision":    rt.ReleaseVersion.Revision,
			"TargetState": rt.TransitionState.String(),
		}).Warn("computed release state")
		rts.Data = append(rts.Data, rt)

		// Add this release tartget to the transaction
		rts.transaction.Versions().AddReleaseVersion(rt.ReleaseVersion)
	}

	// iterate current releases, any that do not exist in rollbackReleaseList set to delete
	for _, rel := range currentReleases {
		if rel.Status == cluster.Status_DELETED {
			log.WithFields(log.Fields{
				"RunningReleaseName": rel.ReleaseName,
			}).Debug("Already deleted")
			// Already deleted, so noop
			continue
		} else {
			log.WithFields(log.Fields{
				"RunningReleaseName": rel.ReleaseName,
				"Status":             rel.Status,
			}).Debug("not deleted")
		}

		// if the current release exists in manifest, move on
		if rts.HasRelease(rel.ReleaseName) {
			continue
		}

		rv := &cluster.Version{
			Name:      rel.ReleaseName,
			Namespace: rel.Namespace,
		}
		rts.Data = append(rts.Data, &ReleaseTarget{
			ReleaseMeta: &cluster.ReleaseMeta{
				ReleaseName: rel.ReleaseName,
				Namespace:   rel.Namespace,
			},
			TransitionState: Deletable,
			ReleaseVersion:  &cluster.Version{},
		})
		rv.SetModified()
	}
	return rts, nil
}

func (rt *ReleaseTargets) HasRelease(releaseName string) bool {
	for _, v := range rt.Data {
		if v.ReleaseMeta.ReleaseName == releaseName {
			return true
		}
	}
	return false
}

func (rt *ReleaseTargets) dryRun(session cluster.Sessioner) error {
	for _, v := range rt.Data {
		v.ReleaseMeta.DryRun = true
		switch v.TransitionState {
		case Installable:
			_, _, _, err := session.InstallRelease(v.ReleaseMeta, rt.ManifestName)
			if err != nil {
				return err
			}
		case Upgradable:
			_, _, err := session.UpgradeRelease(v.ReleaseMeta, rt.ManifestName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rt *ReleaseTargets) Diff(session cluster.Sessioner) (*ReleaseTargets, error) {
	var err error
	for _, v := range rt.Data {
		v.ReleaseMeta.DryRun = true
		switch v.TransitionState {
		case Upgradable:
			v.Changed, v.Diff, err = session.DiffRelease(v.ReleaseMeta)
			if err != nil {
				return nil, err
			}
		}
	}
	return rt, nil
}

func (rt *ReleaseTargets) LogDiff() {
	for _, v := range rt.Data {
		switch v.TransitionState {
		case Installable:
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

func (rt *ReleaseTargets) Apply(opt *CmdOptions) error {

	for _, v := range rt.Data {
		v.ReleaseMeta.DryRun = false
		v.ReleaseMeta.InstallTimeout = 120
		log.WithFields(log.Fields{
			"ReleaseName": v.ReleaseMeta.ReleaseName,
			"Revision":    v.ReleaseVersion.Revision,
			"Transition":  v.TransitionState.String(),
		}).Warn("Applying transition to release")
		switch v.TransitionState {
		case Installable, Replaceable:
			if err := func() error {
				//This closure removes a "break OUT"
				var innerErr error
				if v.TransitionState == Replaceable {
					//The release exists, it needs to be deleted
					dm := &cluster.DeleteMeta{
						ReleaseName:   v.ReleaseMeta.ReleaseName,
						Namespace:     v.ReleaseMeta.Namespace,
						DeleteTimeout: v.ReleaseMeta.InstallTimeout,
					}
					log.WithFields(log.Fields{
						"Name":        v.ReleaseMeta.ReleaseName,
						"Namespace":   v.ReleaseMeta.Namespace,
						"InstallWait": v.ReleaseMeta.InstallWait,
					}).Info("Deleting (force install)")
					if err := rt.session.DeleteRelease(dm); err != nil {
						return errors.Wrap(err, "error deleting release before install (forced)")
					}
					v.ReleaseVersion.SetModified()
				}
				log.WithFields(log.Fields{
					"Name":        v.ReleaseMeta.ReleaseName,
					"Namespace":   v.ReleaseMeta.Namespace,
					"InstallWait": v.ReleaseMeta.InstallWait,
				}).Info("Installing")
				for i := 0; i < opt.InstallRetry; i++ {
					msg, relName, relVersion, err := rt.session.InstallRelease(v.ReleaseMeta, rt.ManifestName)
					if err != nil {
						log.WithFields(log.Fields{
							"Name":        v.ReleaseMeta.ReleaseName,
							"Namespace":   v.ReleaseMeta.Namespace,
							"InstallWait": v.ReleaseMeta.InstallWait,
							"Error":       err.Error(),
						}).Debug("Install reported error")
						innerErr = err
						//The state has changed underneath us, but the release needs installed anyhow
						//So delete and try again
						dm := &cluster.DeleteMeta{
							ReleaseName:   v.ReleaseMeta.ReleaseName,
							Namespace:     v.ReleaseMeta.Namespace,
							DeleteTimeout: v.ReleaseMeta.InstallTimeout,
						}
						log.WithFields(log.Fields{
							"Name":        v.ReleaseMeta.ReleaseName,
							"Namespace":   v.ReleaseMeta.Namespace,
							"InstallWait": v.ReleaseMeta.InstallWait,
						}).Info("Deleting (state change)")
						if err := rt.session.DeleteRelease(dm); err != nil {
							//deleting kube-proxy or other connection issues can trigger this, don't abort the retry
							log.Debug(err, "error deleting release before install (forced)")
						}
						/////
						select {
						default:
							_ = <-time.After(1 * time.Second)
						}
						continue
					}
					log.WithFields(log.Fields{
						"Name":        v.ReleaseMeta.ReleaseName,
						"Namespace":   v.ReleaseMeta.Namespace,
						"InstallWait": v.ReleaseMeta.InstallWait,
						"Release":     relName,
						"Version":     relVersion,
					}).Info(msg)
					v.ReleaseVersion.SetRevision(relVersion)
					innerErr = nil
					return nil
				}
				return errors.WithFields(errors.Fields{
					"Name":        v.ReleaseMeta.ReleaseName,
					"Namespace":   v.ReleaseMeta.Namespace,
					"InstallWait": v.ReleaseMeta.InstallWait,
				}).Wrap(innerErr, "Error while installing release")
			}(); err != nil {
				return err
			}

		case Upgradable, Undeleteable:
			if !v.Changed && v.TransitionState != Undeleteable {
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Skipping due to no change")
				// transaction, merge previous forward
				continue
			}
			if v.TransitionState == Undeleteable {
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
					"Revision":  v.ReleaseMeta.Revision,
				}).Info("Rollback before Upgrade (undelete)")
				_, err := rt.session.RollbackRelease(&cluster.RollbackMeta{
					ReleaseName: v.ReleaseVersion.Name,
					Revision:    v.ReleaseVersion.Revision,
				})
				if err != nil {
					return errors.Wrap(err, "Rollback of release failed")
				}
			}
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Upgrading")
			msg, relVersion, err := rt.session.UpgradeRelease(v.ReleaseMeta, rt.ManifestName)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Wrap(err, "error while upgrading release")
			}
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
				"Version":   relVersion,
			}).Info(msg)
			v.ReleaseVersion.SetRevision(relVersion)
		case Deletable:
			//The release exists, it needs to be deleted
			dm := &cluster.DeleteMeta{
				ReleaseName:   v.ReleaseMeta.ReleaseName,
				Namespace:     v.ReleaseMeta.Namespace,
				DeleteTimeout: v.ReleaseMeta.InstallTimeout,
			}
			log.WithFields(log.Fields{
				"Name":        v.ReleaseMeta.ReleaseName,
				"Namespace":   v.ReleaseMeta.Namespace,
				"InstallWait": v.ReleaseMeta.InstallWait,
			}).Info("Deleting (removed from manifest)")
			if err := rt.session.DeleteRelease(dm); err != nil {
				return errors.Wrap(err, "error deleting release before install (forced)")
			}
			v.ReleaseVersion.SetModified()

		default:
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Skipping")
		}

	}
	return nil
}
