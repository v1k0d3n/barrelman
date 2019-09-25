package cmd

import (
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
)

func newVersionCmd(cmd *barrelman.VersionCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:           "version",
		Short:         "Display version and build information",
		Long:          `Display version and build information supplied at build time.`,
		Example:       "barrelman version",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {

			return cmd.Run()
		},
	}
	return cobraCmd
}
