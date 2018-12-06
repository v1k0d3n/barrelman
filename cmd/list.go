package cmd

import (
	"os"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
)

type listCmd struct {
	Options *cmdOptions
	Config  *Config
}

func newListCmd(cmd *listCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "list",
		Short: "apply something",
		Long:  `Something something else...`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			if err := cmd.Run(cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)); err != nil {
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
	return cobraCmd
}

func (cmd *listCmd) Run(session cluster.Sessioner) error {
	var err error
	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}
	log.WithFields(log.Fields{"file": cmd.Options.ConfigFile}).Info("Using config")

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
	list, err := session.Releases()
	if err != nil {
		return errors.Wrap(err, "Failed to get releases")
	}
	for k, v := range list {
		log.WithFields(log.Fields{
			"key":       k,
			"Name":      v.Name,
			"Namespace": v.Namespace,
		}).Warn("Meta")
	}
	return nil
}
