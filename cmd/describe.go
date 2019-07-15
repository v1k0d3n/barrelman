package cmd

import (
	"github.com/charter-oss/barrelman/cmd/util"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

func newDescribeCmd(cmd *barrelman.DescribeCmd) *cobra.Command {

	longDesc := `Display release information stored in a Barrelman manifest version.`

	shortDesc := `Display release information in a Barrelman manifest version.`

	examples := `barrelman describe lamp-stack 5`

	cobraCmd := &cobra.Command{
		Use:           "describe [manifest name] [version]",
		Short:         shortDesc,
		Long:          longDesc,
		Example:       examples,
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.ManifestName = args[0]

			verTmp, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Wrap(err, "Failed to parse version from second arguement")
			}
			cmd.ManifestVersion = int32(verTmp)

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

	cmd.LogOptions = cobraCmd.Flags().StringSliceP(
		"log",
		"l",
		nil,
		"Set the log level. [ debug | info | warn | error ]")

	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeConfigFile,
		"kubeconfig",
		util.Default().KubeConfigFile,
		"Set the Kubernetes config file to use for connecting to the cluster.")
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeContext,
		"kubecontext",
		util.Default().KubeContext,
		"use alternate kube context")
	return cobraCmd
}
