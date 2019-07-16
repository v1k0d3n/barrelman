package barrelman

import (
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/cirrocloud/structured/log"
)

type VersionCmd struct{}

func (cmd *VersionCmd) Run() error {
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	return nil
}
