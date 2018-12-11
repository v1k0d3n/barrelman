package cmd

import (
	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	Options *cmdOptions
	Config  *Config
}

func newDeleteCmd(cmd *deleteCmd) *cobra.Command {

	cobraCmd := &cobra.Command{
		Use:   "delete [manifest.yaml]",
		Short: "delete something",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeConfigFile,
		"kube-config",
		Default().KubeConfigFile,
		"use alternate kube config file")
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeContext,
		"kube-context",
		Default().KubeContext,
		"use alternate kube context")
	return cobraCmd
}

func (cmd *deleteCmd) Run() error {
	log.Warn("Barrelman Delete Engage!")

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
	session := cluster.NewSession(cmd.Options.KubeContext, cmd.Options.KubeConfigFile)
	if err = session.Init(); err != nil {
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

	log.Info("syncronizing with remote chart repositories")
	if err := mfest.Sync(cmd.Config.Account); err != nil {
		return errors.Wrap(err, "error while downloading charts")
	}

	if err := DeleteByManifest(mfest, session); err != nil {
		return errors.Wrap(err, "failed to delete by manifest")
	}
	return nil
}

func DeleteByManifest(bm *manifest.Manifest, c *cluster.Session) error {
	deleteList := make(map[string]*cluster.DeleteMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return errors.Wrap(err, "error resolving chart groups")
	}

	releases, err := c.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		deleteList[v.Chart.Metadata.Name] = &cluster.DeleteMeta{
			Name:      v.Name,
			Namespace: "",
		}
	}

	for _, cg := range groups {
		charts, err := bm.GetChartsByName(cg.Data.ChartGroup)
		if err != nil {
			return errors.Wrap(err, "error resolving charts")
		}
		for _, v := range charts {
			if dm, exists := deleteList[v.Name]; exists {
				log.WithFields(log.Fields{
					"Name":    v.Name,
					"Release": dm.Name,
				}).Info("deleting release")
				if err := c.DeleteRelease(dm); err != nil {
					return errors.Wrap(err, "error deleting list")
				}
			}
		}
	}
	return nil
}
