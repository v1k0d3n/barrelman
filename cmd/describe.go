package cmd

import (
	"strconv"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

func newDescribeCmd(cmd *barrelman.DescribeCmd) *cobra.Command {

	longDesc := dedent.Dedent(`
		describe [manifest name] [version]
			Display release information stored in a Barrelman manifest version.
	`)

	shortDesc := dedent.Dedent(`
		display release information in a Barrelman manifest version
	`)

	cobraCmd := &cobra.Command{
		Use:   "describe [manifest name] [version]",
		Short: shortDesc,
		Long:  longDesc,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("command requires 'manifest name' and 'version'")
			}

			cmd.ManifestName = args[0]

			verTmp, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Wrap(err, "Failed to parse version from second arguement")
			}
			cmd.ManifestVersion = int32(verTmp)

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

	cmd.LogOptions = cobraCmd.Flags().StringSliceP(
		"log",
		"l",
		nil,
		"Set the log level. [ debug | info | warn | error ]")

	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeConfigFile,
		"kubeconfig",
		Default().KubeConfigFile,
		"Set the Kubernetes config file to use for connecting to the cluster.")
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeContext,
		"kubecontext",
		Default().KubeContext,
		"use alternate kube context")
	return cobraCmd
}
