package cmd

import (
	"fmt"
	"os"

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
		Run: func(cobraCmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			if err := cmd.Run(); err != nil {
				log.Error(err)
				os.Exit(1)
			}
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

func (cmd *applyCmd) Run() error {
	var err error
	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}
	log.WithFields(log.Fields{"file": cmd.Options.ConfigFile}).Info("Using config")

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	// Open connections to the k8s APIs
	session, err := cluster.NewSession(cmd.Options.KubeContext, cmd.Options.KubeConfigFile)
	if err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}
	log.WithFields(log.Fields{
		"file": session.KubeConfig,
	}).Info("Using kube config")
	if session.KubeContext != "" {
		log.WithFields(log.Fields{
			"file": session.KubeContext,
		}).Info("Using kube context")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	if !cmd.Options.NoSync {
		log.Info("syncronizing with remote chart repositories")
		if err := mfest.Sync(cmd.Config.Account); err != nil {
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
		rt.LogDiff(session)
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
	for _, r := range *cmd.Options.Force {
		if r == rel.Name {
			return true
		}
	}
	return false
}
func (cmd *applyCmd) ComputeReleases(
	session *cluster.Session,
	archives *manifest.ArchiveFiles,
	currentReleases map[string]*cluster.ReleaseMeta) releaseTargets {
	rt := releaseTargets{}

	for _, v := range archives.List {
		if rel, ok := currentReleases[v.Name]; ok {
			if cmd.isInForce(rel) {
				rt = append(rt,
					&releaseTarget{
						State: Replaceable,
						ReleaseMeta: &cluster.ReleaseMeta{
							Path:           v.Path,
							Name:           rel.Name,
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
							Name:           rel.Name,
							Namespace:      v.Namespace,
							ValueOverrides: v.Overrides,
						},
					})
			}
		} else {
			rt = append(rt,
				&releaseTarget{
					State: Installable,
					ReleaseMeta: &cluster.ReleaseMeta{
						Path:           v.Path,
						Name:           v.Name,
						Namespace:      v.Namespace,
						ValueOverrides: v.Overrides,
					},
				})
		}
	}
	return rt
}

func (rt releaseTargets) dryRun(session *cluster.Session) error {
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

func (rt releaseTargets) Diff(session *cluster.Session) (releaseTargets, error) {
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

func (rt releaseTargets) LogDiff(session *cluster.Session) {
	for _, v := range rt {
		switch v.State {
		case Installable:
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.Name,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info("Would install")
		case Upgradable:
			if v.Changed {
				log.WithFields(log.Fields{
					"Name": v.ReleaseMeta.Name,
				}).Info("Diff")
				//Print the byte content and keep formatting, its fancy
				fmt.Printf("----%v\n%v_____\n", v.ReleaseMeta.Name, string(v.Diff))
			} else {
				log.WithFields(log.Fields{
					"Name": v.ReleaseMeta.Name,
				}).Info("No change")
			}
		}
	}
}

func (rt releaseTargets) Apply(session *cluster.Session, opt *cmdOptions) error {
	for _, v := range rt {
		v.ReleaseMeta.DryRun = false
		switch v.State {

		case Installable, Replaceable:
			if err := func() error {
				//This closure removes a "break OUT"
				var err error
				if v.State == Replaceable {
					//The relase exists, it needs to be deleted
					dm := &cluster.DeleteMeta{
						Name:      v.ReleaseMeta.Name,
						Namespace: v.ReleaseMeta.Namespace,
					}
					log.WithFields(log.Fields{
						"Name":      v.ReleaseMeta.Name,
						"Namespace": v.ReleaseMeta.Namespace,
					}).Info("Deleting (force install)")
					if err := session.DeleteRelease(dm); err != nil {
						return errors.Wrap(err, "error deleting release before install (forced)")
					}
				}
				for i := 0; i < opt.InstallRetry; i++ {
					msg, relName, err := session.InstallRelease(v.ReleaseMeta, []byte{})
					if err != nil {
						continue
					}
					log.WithFields(log.Fields{
						"Name":      v.ReleaseMeta.Name,
						"Namespace": v.ReleaseMeta.Namespace,
						"Release":   relName,
					}).Info(msg)
					return nil
				}
				return errors.WithFields(errors.Fields{
					"Name":      v.ReleaseMeta.Name,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Wrap(err, "Error while installing release")
			}(); err != nil {
				return err
			}

		case Upgradable:
			if !v.Changed {
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.Name,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Skipping")
				continue
			}
			msg, err := session.UpgradeRelease(v.ReleaseMeta)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.ReleaseMeta.Name,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Wrap(err, "error while upgrading release")
			}
			log.WithFields(log.Fields{
				"Name":      v.ReleaseMeta.Name,
				"Namespace": v.ReleaseMeta.Namespace,
			}).Info(msg)
		}
	}
	return nil
}
