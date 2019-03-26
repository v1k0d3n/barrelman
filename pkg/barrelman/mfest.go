package barrelman

import (
	"github.com/cirrocloud/yamlpack"

	"github.com/charter-se/barrelman/pkg/manifest"
	"github.com/charter-se/structured/errors"
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
	return archives, err
}

func processManifestSections(config *manifest.Config, ys []*yamlpack.YamlSection, noSync bool) (*manifest.ArchiveFiles, error) {
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
	return archives, err
}
