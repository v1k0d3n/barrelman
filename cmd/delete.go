package cmd

import (
	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/barrelman/version"
	"github.com/charter-se/structured"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	Options    *cmdOptions
	Config     *Config
	Log        structured.Logger
	LogOptions *[]string
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
			cmd.Log = log.New(logSettings(cmd.LogOptions)...)
			session := cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)
			if err := cmd.Run(session); err != nil {
				return err
			}
			return nil
		},
	}
	cobraCmd.Flags().BoolVar(
		&cmd.Options.NoSync,
		"nosync",
		false,
		"disable remote sync")
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
	return cobraCmd
}

func (cmd *deleteCmd) Run(session cluster.Sessioner) error {
	var err error

	ver := version.Get()
	cmd.Log.WithFields(log.Fields{
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

	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	if session.GetKubeConfig() != "" {
		cmd.Log.WithFields(log.Fields{
			"file": session.GetKubeConfig(),
		}).Info("Using kube config")
	}
	if session.GetKubeContext() != "" {
		cmd.Log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
		Log:          cmd.Log,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	if !cmd.Options.NoSync {
		if err := mfest.Sync(); err != nil {
			return errors.Wrap(err, "error while downloading charts")
		}
	}

	if err := DeleteByManifest(cmd.Log, mfest, session); err != nil {
		return errors.Wrap(err, "failed to delete by manifest")
	}
	return nil
}

func DeleteByManifest(logger structured.Logger, bm *manifest.Manifest, session cluster.Sessioner) error {
	deleteList := make(map[string]*cluster.DeleteMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return errors.Wrap(err, "error resolving chart groups")
	}

	releases, err := session.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		deleteList[v.Chart.GetMetadata().Name] = &cluster.DeleteMeta{
			ReleaseName: v.ReleaseName,
			Namespace:   "",
		}
	}

	for _, cg := range groups {
		charts, err := bm.GetChartsByChartName(cg.Data.ChartGroup)
		if err != nil {
			return errors.Wrap(err, "error resolving charts")
		}
		for _, v := range charts {
			for _, rel := range deleteList {
				if rel.ReleaseName == v.Data.ReleaseName {
					//if dm, exists := deleteList[v.Data.ReleaseName]; exists {
					logger.WithFields(log.Fields{
						"Name":    v.Metadata.Name,
						"Release": rel.ReleaseName,
					}).Info("deleting release")
					if err := session.DeleteRelease(rel); err != nil {
						return errors.Wrap(err, "error deleting list")
					}
				}
			}
		}
	}
	return nil
}
