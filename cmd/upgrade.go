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

type upgradeCmd struct {
	Options *cmdOptions
	Config  *Config
}

func newUpgradeCmd(cmd *upgradeCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "upgrade [manifest.yaml]",
		Short: "upgrade something",
		Long:  `Something something else...`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			if err := runUpgradeCmd(cmd); err != nil {
				log.Error(err)
				os.Exit(1)
			}
		},
	}
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

	return cobraCmd
}

func runUpgradeCmd(cmd *upgradeCmd) error {
	log.Warn("Barrelman Upgrade Engage!")

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
	c, err := cluster.NewSession(Default().KubeConfigFile)
	if err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	log.Info("syncronizing with remote chart repositories")
	if err := mfest.Sync(cmd.Config.Account); err != nil {
		return errors.Wrap(err, "error while downloading charts")
	}

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

	err = UpgradeByManifest(cmd, mfest, c)
	if err != nil {
		return errors.Wrap(err, "error while upgrading release")
	}
	return nil
}

func UpgradeByManifest(cmd *upgradeCmd, bm *manifest.Manifest, c *cluster.Session) error {
	upgradeList := make(map[string]*cluster.ReleaseMeta)

	releases, err := c.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		upgradeList[v.Chart.Metadata.Name] = &cluster.ReleaseMeta{
			Name:      v.Name,
			Namespace: v.Namespace,
		}
	}

	archives, err := bm.CreateArchives()
	if err != nil {
		log.Error(errors.Wrap(err, "failed to create archives"))
		os.Exit(1)
	}
	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()

	err = doUpgrade(cmd, c, archives, upgradeList)
	if err != nil {
		return errors.Wrap(err, "Manifest upgrade failed")
	}

	return nil
}

func doUpgrade(cmd *upgradeCmd,
	c *cluster.Session,
	archives *manifest.ArchiveFiles,
	upgradeList map[string]*cluster.ReleaseMeta) error {

	var dryRun bool
	if cmd.Options.DryRun || cmd.Options.Diff {
		dryRun = true
	}

	for _, v := range archives.List {
		if rel, exists := upgradeList[v.Name]; exists {
			upgradeList[v.Name].Path = v.Path
			msg, res, err := c.UpgradeRelease(&cluster.ReleaseMeta{
				Path:           v.Path,
				Name:           rel.Name,
				Namespace:      v.Namespace,
				ValueOverrides: v.Overrides,
				DryRun:         dryRun,
			})
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.Name,
					"Namespace": v.Namespace,
					"Release":   rel.Name,
				}).Wrap(err, "error while installing release")
			}
			log.WithFields(log.Fields{
				"Name":      v.Name,
				"Namespace": v.Namespace,
				"Release":   rel.Name,
				"Results":   msg,
			}).Info("upgraded release")
			if cmd.Options.Diff {
				fmt.Printf("Diff: %v<<*******\n", res)
			}
		} else {
			relName, res, err := c.InstallRelease(&cluster.ReleaseMeta{
				Path:           v.Path,
				Namespace:      v.Namespace,
				Name:           v.Name,
				ValueOverrides: v.Overrides,
				DryRun:         dryRun,
			}, []byte{})
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.Name,
					"Namespace": v.Namespace,
				}).Wrap(err, "error while installing release")
			}
			log.WithFields(log.Fields{
				"Name":      v.Name,
				"Namespace": v.Namespace,
				"Release":   relName,
				"Overrides": string(v.Overrides),
			}).Info("installed release")
			if cmd.Options.Diff {
				fmt.Printf("Diff: %v<<*******\n", res)
			}
		}
	}
	return nil
}
