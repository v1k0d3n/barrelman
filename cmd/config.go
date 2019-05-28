package cmd

import (
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
)


func newConfigCmd(cmd *barrelman.ConfigCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "config",
		Short: "get default config file",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.Run()
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
