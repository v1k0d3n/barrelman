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

			if err := cmd.Run(cmd.Options.KubeConfigFile); err!=nil {
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

	return cobraCmd
}
