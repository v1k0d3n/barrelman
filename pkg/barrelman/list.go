package barrelman

import (
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/cirrocloud/structured"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

type ListCmd struct {
	Options      *CmdOptions
	Config       *Config
	ManifestName string
	Log          structured.Logger
	LogOptions   *[]string
}

func (cmd *ListCmd) Run(session cluster.Sessioner) error {
	var err error
	log.Rep(version.Get()).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	log.Debug("connecting to cluster")
	if err := session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	// No name supplied, list all top level Barrelman manifests
	if cmd.ManifestName == "" {
		list, err := session.ListManifests()
		if err != nil {
			return errors.Wrap(err, "Failed to get releases")
		}
		for _, v := range list {
			log.Rep(v).Info("Barrelman Manifest")
		}
		return nil
	}

	// Name was supplied, list the releases
	list, err := session.ReleasesByManifest(cmd.ManifestName)
	if err != nil {
		return errors.Wrap(err, "Failed to get releases")
	}
	for _, v := range list {
		log.WithFields(log.Fields{
			"Name":      v.ReleaseName,
			"Revision":  v.Revision,
			"Namespace": v.Namespace,
		}).Info("Release in Manifest")
	}

	return nil
}
