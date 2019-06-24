package cmd

import (
	"os"
	"strconv"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

func newRollbackCmd(cmd *barrelman.RollbackCmd) *cobra.Command {

	longDesc := dedent.Dedent(`
		barrelman rollback [manifest name] [manifest version]
			Sets release versions to those recorded in a Barrelman rollback manifest.

			Rollback manifests are saved in the cluster when barrelman sucessfully commits a change
			to the manifest release group.
	`)

	shortDesc := dedent.Dedent(`
		set release versions to a previous Barrelman save state 
	`)

	cobraCmd := &cobra.Command{
		Use:   "rollback [manifest name] [manifest version]",
		Short: shortDesc,
		Long:  longDesc,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("rollback requires 'manifest name' and 'version")
			}
			cmd.ManifestName = args[0]

			verTmp, err := strconv.Atoi(args[1])
			if err != nil {
				log.Error(errors.Wrap(err, "Failed to parse version from second arguement"))
				os.Exit(1)
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
