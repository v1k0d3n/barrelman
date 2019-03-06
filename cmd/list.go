package cmd

import (
	"github.com/spf13/cobra"

	"github.com/charter-se/barrelman/pkg/barrelman"
	"github.com/charter-se/barrelman/pkg/cluster"
	"github.com/charter-se/structured/log"
)

func newListCmd(cmd *barrelman.ListCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "list",
		Short: "apply something",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			log.Configure(logSettings(cmd.LogOptions)...)
			session := cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)
			if err := cmd.Run(session); err != nil {
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
	return cobraCmd
}
