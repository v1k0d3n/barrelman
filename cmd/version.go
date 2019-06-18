package cmd

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
)

func newVersionCmd(cmd *barrelman.VersionCmd) *cobra.Command {
	longDesc := dedent.Dedent(`
		barrelman version
			show build version including git branch and commit tag
	`)

	shortDesc := dedent.Dedent(`
		show version information
	`)

	cobraCmd := &cobra.Command{
		Use:   "version",
		Short: shortDesc,
		Long:  longDesc,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true

			return cmd.Run()
		},
	}
	return cobraCmd
}
