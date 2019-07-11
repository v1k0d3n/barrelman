package cmd

import (
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/cirrocloud/structured/log"
)

func newDeleteCmd(cmd *barrelman.DeleteCmd) *cobra.Command {

	longDesc := strings.TrimSpace(dedent.Dedent(`
		The delete command deletes all releases in a manifest.

		All releases currently deployed in the matching manifest will be deleted, 
		as will all releases currently configured in the supplied manifest file.
		
		When no manifest is specified Barrelman will attempt to delete ~/.barrelman/manifest.yaml.
	`))

	shortDesc := `Delete all releases configured in the manifest.`

	examples := `barrelman delete lamp-stack.yaml`

	cobraCmd := &cobra.Command{
		Use:           "barrelman delete [manifest-file]",
		Short:         shortDesc,
		Long:          longDesc,
		Args:          cobra.ExactArgs(1),
		Example:       examples,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			cmd.Options.ManifestFile = args[0]
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
		&cmd.Options.NoSync,
		"nosync",
		false,
		"disable remote sync")
	return cobraCmd
}
