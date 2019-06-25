package cmd

import (
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/cirrocloud/structured/log"
)

func newApplyCmd(cmd *barrelman.ApplyCmd) *cobra.Command {
	longDesc := strings.TrimSpace(dedent.Dedent(`
		The apply command attempts to commit the differences between a supplied manifest and
		the current Kubernetes cluster state.

		For the given manifest each chart is installed on the Kubernetes cluster.
	`))

	shortDesc := `Apply the given manifest to the cluster.`

	examples := `barrelman apply lamp-stack.yaml`

	cobraCmd := &cobra.Command{
		Use:           "apply [manifest.yaml]",
		Short:         shortDesc,
		Long:          longDesc,
		Example:       examples,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.Options.ManifestFile = args[0]

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
