package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

func newHistoryCmd(cmd *barrelman.HistoryCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "history [manifest.yaml]",
		Short: "history something",
		Long:  `history something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.ManifestName = args[0]
			}
			if len(args) > 1 {

				verTmp, err := strconv.Atoi(args[1])
				if err != nil {
					log.Error(errors.Wrap(err, "Failed to parse version from second arguement"))
					os.Exit(1)
				}
				cmd.ManifestVersion = int32(verTmp)
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

	cmd.LogOptions = cobraCmd.Flags().StringSliceP(
		"log",
		"l",
		nil,
		"log options (e.g. --log=debug,JSON")

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
	return cobraCmd
}
