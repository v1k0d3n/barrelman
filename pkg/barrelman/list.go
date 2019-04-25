package barrelman

import (
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

type ListCmd struct {
	Options    *CmdOptions
	Config     *Config
	Log        structured.Logger
	LogOptions *[]string
}

func (cmd *ListCmd) Run(session cluster.Sessioner) error {
	var err error
	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	if session.GetKubeConfig() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeConfig(),
		}).Info("Using kube config")
	}
	if session.GetKubeContext() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}
	list, err := session.Releases()
	if err != nil {
		return errors.Wrap(err, "Failed to get releases")
	}
	for k, v := range list {
		log.WithFields(log.Fields{
			"key":       k,
			"Name":      v.ReleaseName,
			"Namespace": v.Namespace,
		}).Warn("Meta")
	}
	return nil
}
