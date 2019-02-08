package cmd

import (
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
)

func processManifest(config *manifest.Config, noSync bool) (*manifest.ArchiveFiles, error) {
	// Open and initialize the manifest
	mfest, err := manifest.New(config)
	if err != nil {
		return nil, errors.Wrap(err, "error while initializing manifest")
	}

	if !noSync {
		if err := mfest.Sync(); err != nil {
			return nil, errors.Wrap(err, "error while downloading charts")
		}
	}

	//Build/update chart archives from manifest
	archives, err := mfest.CreateArchives()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create archives")
	}
	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()
	return archives, err
}
