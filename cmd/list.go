package cmd

import (
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/log"
)

func newListCmd(cmd *barrelman.ListCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "list [manifest name]",
		Short: "list something",
		Long:  `lists the current releases associated with the provided manifest name`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.ManifestName = args[0]
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

	f := cobraCmd.Flags()
	cmd.LogOptions = f.StringSliceP(
		"log",
		"l",
		nil,
		"log options (e.g. --log=debug,JSON")
	f.StringArrayVarP(
		cmd.ManifestNames,
		"manifests",
		"m",
		[]string{},
		"list releases by manifest name(s)")
	f.StringVar(
		&cmd.Options.KubeConfigFile,
		"kubeconfig",
		Default().KubeConfigFile,
		"use alternate kube config file")
	f.StringVar(
		&cmd.Options.KubeContext,
		"kubecontext",
		Default().KubeContext,
		"use alternate kube context")
	return cobraCmd
}
