package cmd

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/log"
)

func newListCmd(cmd *barrelman.ListCmd) *cobra.Command {
	longDesc := dedent.Dedent(`
		barrelman list
			show all Barrelman manifests stored in kubernetes cluster
			
		barrelman list [manifest name]
			show releases and versions asssociated with a Barrelman manifest currently running in a Kubernetes cluster
	`)

	shortDesc := dedent.Dedent(`
		list manifests and releases
	`)

	cobraCmd := &cobra.Command{
		Use:   "list",
		Short: shortDesc,
		Long:  longDesc,
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
