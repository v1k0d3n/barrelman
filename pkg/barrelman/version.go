package barrelman

import (
	"fmt"

	"github.com/charter-oss/barrelman/pkg/version"
)

type VersionCmd struct{}

var templ = []byte(`Version: {{.Version}}
Branch: {{.Branch}}
Commit: {{.Commit}}`)

func (cmd *VersionCmd) Run() error {
	ver := version.Get()
	fmt.Printf("\nBarrelman deployment tool\n\n")
	fmt.Printf("\tVersion: %v\n", ver.Version)
	fmt.Printf("\tBranch: %v\n", ver.Branch)
	fmt.Printf("\tCommit: %v\n", ver.Commit)

	return nil
}
