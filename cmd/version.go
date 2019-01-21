package cmd

import (
	"fmt"

	"github.com/charter-se/barrelman/version"
	"github.com/spf13/cobra"
)

type versionCmd struct{}

var templ = []byte(`Version: {{.Version}}
Branch: {{.Branch}}
Commit: {{.Commit}}`)

func newVersionCmd(cmd *versionCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "version",
		Short: "version something",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true

			return cmd.Run()
		},
	}
	return cobraCmd
}

func (cmd *versionCmd) Run() error {
	ver := version.Get()
	fmt.Printf("\nBarrelman deployment tool\n\n")
	fmt.Printf("\tVersion: %v\n", ver.Version)
	fmt.Printf("\tBranch: %v\n", ver.Branch)
	fmt.Printf("\tCommit: %v\n", ver.Commit)

	return nil
}
