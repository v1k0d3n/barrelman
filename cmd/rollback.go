package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

func newRollbackCmd(cmd *barrelman.RollbackCmd) *cobra.Command {

	longDesc := strings.TrimSpace(dedent.Dedent(`
		Rollback sets release versions to those recorded in a Barrelman rollback manifest.
		
		Rollback manifests are saved in the cluster when barrelman sucessfully commits a change
		to the manifest release group.`))

	shortDesc := `Set release versions to a previous Barrelman save state.`

	examples := `barrelman rollback lamp-stack 5`

	cobraCmd := &cobra.Command{
		Use:           "barrelman rollback [manifest-name] [revision]",
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
				log.Error(errors.Wrap(err, "Failed to parse version from second arguement"))
				os.Exit(1)
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
		"log options (e.g. --log=debug,JSON")

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
