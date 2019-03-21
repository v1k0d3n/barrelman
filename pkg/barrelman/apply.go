package barrelman

import (
	"fmt"
	"time"

	"github.com/charter-se/barrelman/pkg/cluster"
	"github.com/charter-se/barrelman/pkg/manifest"
	"github.com/charter-se/barrelman/pkg/version"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
)

const (
	Installable = iota
	Upgradable
	Replaceable
	Deletable
	NoChange
)

type ApplyCmd struct {
	Options    *CmdOptions
	Config     *Config
	LogOptions *[]string
}

type ReleaseTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	Chart       *cluster.Chart
	State       int
	Diff        []byte
	Changed     bool
}

type ReleaseTargets []*ReleaseTarget

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
		return errors.WithFields(errors.Fields{"Dir": cmd.Options.DataDir}).Wrap(err, "failed to create working directory")
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

	archives, err := processManifest(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "apply failed")
	}

	releases, err := session.Releases()
	if err != nil {
		return errors.Wrap(err, "failed to get current releases")
	}

	rt, err := cmd.ComputeReleases(session, archives, releases)
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
	err = rt.Apply(session, cmd.Options)
	if err != nil {
		return errors.Wrap(err, "Manifest upgrade failed")
	}

	return nil
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
//states may be one of 'Installable', 'Upgradeable', or 'Replaceable'
func (cmd *ApplyCmd) ComputeReleases(
	session cluster.Sessioner,
	archives *manifest.ArchiveFiles,
	currentReleases map[string]*cluster.ReleaseMeta) (ReleaseTargets, error) {
	rt := ReleaseTargets{}

	for _, v := range archives.List {
		releaseExists := false
		inChart, err := session.ChartFromArchive(v.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate chart from archive")
		}
		for _, rel := range currentReleases {
			if rel.ReleaseName == v.ReleaseName {
				releaseExists = true
				if cmd.isInForce(rel) || rel.Status == cluster.Status_FAILED {
					rt = append(rt,
						&ReleaseTarget{
							State: Replaceable,
							ReleaseMeta: &cluster.ReleaseMeta{
								Chart:          inChart,
								ReleaseName:    rel.ReleaseName,
								Namespace:      v.Namespace,
								ValueOverrides: v.Overrides,
							},
						})
				} else {
					rt = append(rt,
						&ReleaseTarget{
							State: Upgradable,
							ReleaseMeta: &cluster.ReleaseMeta{
								Chart:          inChart,
								ReleaseName:    rel.ReleaseName,
								Namespace:      v.Namespace,
								ValueOverrides: v.Overrides,
							},
						})
				}
			}
		}
		if !releaseExists {
			rt = append(rt,
				&ReleaseTarget{
					State: Installable,
					ReleaseMeta: &cluster.ReleaseMeta{
						Chart:          inChart,
						ReleaseName:    v.ReleaseName,
						Namespace:      v.Namespace,
						ValueOverrides: v.Overrides,
					},
				})
		}
	}
	return rt, nil
}

func (rt ReleaseTargets) dryRun(session cluster.Sessioner) error {
	for _, v := range rt {
		v.ReleaseMeta.DryRun = true
		switch v.State {
		case Installable:
			_, _, err := session.InstallRelease(v.ReleaseMeta)
			if err != nil {
				return err
			}
		case Upgradable:
			_, err := session.UpgradeRelease(v.ReleaseMeta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rt ReleaseTargets) Diff(session cluster.Sessioner) (ReleaseTargets, error) {
	var err error
	for _, v := range rt {
		v.ReleaseMeta.DryRun = true
		switch v.State {
		case Upgradable:
			v.Changed, v.Diff, err = session.DiffRelease(v.ReleaseMeta)
			if err != nil {
				return nil, err
			}
		}
	}
	return rt, nil
}

func (rt ReleaseTargets) LogDiff() {
	for _, v := range rt {
		switch v.State {
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

func (rt ReleaseTargets) Apply(session cluster.Sessioner, opt *CmdOptions) error {
	for _, v := range rt {
		v.ReleaseMeta.DryRun = false
		v.ReleaseMeta.InstallTimeout = 120
		switch v.State {
		case Installable, Replaceable:
			if err := func() error {
				//This closure removes a "break OUT"

				var innerErr error
				if v.State == Replaceable {
					//The release exists, it needs to be deleted
					dm := &cluster.DeleteMeta{
						ReleaseName:   v.ReleaseMeta.ReleaseName,
						Namespace:     v.ReleaseMeta.Namespace,
						DeleteTimeout: v.ReleaseMeta.InstallTimeout,
					}
					log.WithFields(log.Fields{
						"Name":      v.ReleaseMeta.ReleaseName,
						"Namespace": v.ReleaseMeta.Namespace,
					}).Info("Deleting (force install)")
					if err := session.DeleteRelease(dm); err != nil {
						return errors.Wrap(err, "error deleting release before install (forced)")
					}
				}
				for i := 0; i < opt.InstallRetry; i++ {
					msg, relName, err := session.InstallRelease(v.ReleaseMeta)
					if err != nil {
						log.WithFields(log.Fields{
							"Name":      v.ReleaseMeta.ReleaseName,
							"Namespace": v.ReleaseMeta.Namespace,
							"Error":     err.Error(),
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
							"Name":      v.ReleaseMeta.ReleaseName,
							"Namespace": v.ReleaseMeta.Namespace,
						}).Info("Deleting (state change)")
						if err := session.DeleteRelease(dm); err != nil {
							return errors.Wrap(err, "error deleting release before install (forced)")
						}
						/////
						select {
						default:
							_ = <-time.After(1 * time.Second)
						}
						continue
					}
					log.WithFields(log.Fields{
						"Name":      v.ReleaseMeta.ReleaseName,
						"Namespace": v.ReleaseMeta.Namespace,
						"Release":   relName,
					}).Info(msg)
					innerErr = nil
					return nil
				}
				return errors.WithFields(errors.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Wrap(innerErr, "Error while installing release")
			}(); err != nil {
				return err
			}

		case Upgradable:
			if !v.Changed {
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Skipping")
				continue
			}
			msg, err := session.UpgradeRelease(v.ReleaseMeta)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Wrap(err, "error while upgrading release")
			}
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info(msg)
		default:
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.ReleaseName,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Skipping")
		}

	}
	return nil
}
