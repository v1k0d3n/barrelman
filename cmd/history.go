package cmd

import (
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
)

func newHistoryCmd(cmd *barrelman.HistoryCmd) *cobra.Command {

	longDesc := strings.TrimSpace(dedent.Dedent(`
		Display Barrelman manifest revision history.
		
		A new manifest revision is created upon sucessful change is performed on the cluster.
	`))

	shortDesc := `List manifest revision history.`

	examples := `barrelman history lamp-stack`

	cobraCmd := &cobra.Command{
		Use:           "history [manifest name]",
		Short:         shortDesc,
		Long:          longDesc,
		Example:       examples,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.ManifestName = args[0]

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
	return cobraCmd
}
