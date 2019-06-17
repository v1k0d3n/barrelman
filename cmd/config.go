package cmd

import (
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/log"
)

func newConfigCmd(cmd *barrelman.ConfigCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "config [manifest.yaml]",
		Short: "config something",
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
		"retry install (n) times")

	return cobraCmd
}
