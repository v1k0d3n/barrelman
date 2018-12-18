package cmd

import (
	"fmt"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
)

const (
	Installable = iota
	Upgradable
	Replaceable
	Deletable
	NoChange
)

type applyCmd struct {
	Options *cmdOptions
	Config  *Config
}

type releaseTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	State       int
	Diff        []byte
	Changed     bool
}

type releaseTargets []*releaseTarget

func newApplyCmd(cmd *applyCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "apply [manifest.yaml]",
		Short: "apply something",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			if err := cmd.Run(cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)); err != nil {
				return err
			}
			return nil
		},
	}
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeConfigFile,
		"kubeconfig",
		Default().KubeConfigFile,
		"use alternate kube config file")
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeContext,
		"kubecontext",
		Default().KubeContext,
		"use alternate kube context")
	cobraCmd.Flags().BoolVar(
		&cmd.Options.DryRun,
		"dry-run",
		false,
		"test all charts with server")
	cobraCmd.Flags().BoolVar(
		&cmd.Options.Diff,
		"diff",
		false,
		"Display differences")
	cobraCmd.Flags().BoolVar(
		&cmd.Options.NoSync,
		"nosync",
		false,
		"disable remote sync")
	cmd.Options.Force = cobraCmd.Flags().StringSlice(
		"force",
		*(Default().Force),
		"force apply chart name(s)")
	cobraCmd.Flags().IntVar(
		&cmd.Options.InstallRetry,
		"install-retry",
		Default().InstallRetry,
		"retry install (n) times (Kubernetes bug workaround)")

	return cobraCmd
}

func (cmd *applyCmd) Run(session cluster.Sessioner) error {
	var err error

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}
	log.WithFields(log.Fields{
		"file": session.GetKubeConfig(),
	}).Info("Using kube config")
	if session.GetKubeContext() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	if !cmd.Options.NoSync {
		if err := mfest.Sync(); err != nil {
			return errors.Wrap(err, "error while downloading charts")
		}
	}

	//Build/update chart archives from manifest
	archives, err := mfest.CreateArchives()
	if err != nil {
		return errors.Wrap(err, "failed to create archives")
	}
	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()

	releases, err := session.Releases()
	if err != nil {
		return errors.Wrap(err, "failed to get current releases")
	}

	rt := cmd.ComputeReleases(session, archives, releases)

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
	log.Info("Doing apply")
	err = rt.Apply(session, cmd.Options)
	if err != nil {
		return errors.Wrap(err, "Manifest upgrade failed")
	}

	return nil
}

//IsReplaceable checks a release against the --force flag values to see if an existing release should be replaced via delete
func (cmd *applyCmd) isInForce(rel *cluster.ReleaseMeta) bool {
	//Checks for releases configured for Force by cmdline
	for _, r := range *cmd.Options.Force {
		if r == rel.MetaName {
			return true
		}
	}
	return false
}

func (cmd *applyCmd) ComputeReleases(
	session cluster.Sessioner,
	archives *manifest.ArchiveFiles,
	currentReleases map[string]*cluster.ReleaseMeta) releaseTargets {
	rt := releaseTargets{}

	for _, v := range archives.List {
		releaseExists := false
		for _, rel := range currentReleases {
			if rel.ChartName == v.ChartName {
				releaseExists = true
				if cmd.isInForce(rel) || rel.Status == cluster.Status_FAILED {
					rt = append(rt,
						&releaseTarget{
							State: Replaceable,
							ReleaseMeta: &cluster.ReleaseMeta{
								Path:           v.Path,
								ReleaseName:    rel.ReleaseName,
								Namespace:      v.Namespace,
								ValueOverrides: v.Overrides,
							},
						})
				} else {
					rt = append(rt,
						&releaseTarget{
							State: Upgradable,
							ReleaseMeta: &cluster.ReleaseMeta{
								Path:           v.Path,
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
				&releaseTarget{
					State: Installable,
					ReleaseMeta: &cluster.ReleaseMeta{
						Path:           v.Path,
						ReleaseName:    v.ReleaseName,
						Namespace:      v.Namespace,
						ValueOverrides: v.Overrides,
					},
				})
		}
	}
	return rt
}

func (rt releaseTargets) dryRun(session cluster.Sessioner) error {
	for _, v := range rt {
		v.ReleaseMeta.DryRun = true
		switch v.State {
		case Installable:
			_, _, err := session.InstallRelease(v.ReleaseMeta, []byte{})
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

func (rt releaseTargets) Diff(session cluster.Sessioner) (releaseTargets, error) {
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

func (rt releaseTargets) LogDiff() {
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

func (rt releaseTargets) Apply(session cluster.Sessioner, opt *cmdOptions) error {
	for _, v := range rt {
		v.ReleaseMeta.DryRun = false
		switch v.State {

		case Installable, Replaceable:
			if err := func() error {
				//This closure removes a "break OUT"

				var innerErr error
				if v.State == Replaceable {
					//The release exists, it needs to be deleted
					dm := &cluster.DeleteMeta{
						ReleaseName: v.ReleaseMeta.ReleaseName,
						Namespace:   v.ReleaseMeta.Namespace,
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
					msg, relName, err := session.InstallRelease(v.ReleaseMeta, []byte{})
					if err != nil {
						innerErr = err
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
		}
	}
	return nil
}
