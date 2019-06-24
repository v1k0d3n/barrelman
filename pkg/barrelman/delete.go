package barrelman

import (
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

type DeleteCmd struct {
	Options    *CmdOptions
	Config     *Config
	LogOptions *[]string
}

func (cmd *DeleteCmd) Run(session cluster.Sessioner) error {
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

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	if !cmd.Options.NoSync {
		if err := mfest.Sync(); err != nil {
			return errors.Wrap(err, "error while downloading charts")
		}
	}

	if err := DeleteByManifest(mfest, session); err != nil {
		return errors.Wrap(err, "failed to delete by manifest")
	}
	return nil
}

func DeleteByManifest(bm *manifest.Manifest, session cluster.Sessioner) error {
	deleteList := make(map[string]*cluster.DeleteMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return errors.Wrap(err, "error resolving chart groups")
	}

	releases, err := session.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		deleteList[v.ReleaseName] = &cluster.DeleteMeta{
			ReleaseName: v.ReleaseName,
			Namespace:   "",
		}
	}

	for _, cg := range groups {
		charts, err := bm.GetChartsByChartName(cg.Data.ChartGroup)
		if err != nil {
			return errors.Wrap(err, "error resolving charts")
		}
		for _, v := range charts {
			for _, rel := range deleteList {
				if rel.ReleaseName == v.Data.ReleaseName {
					//if dm, exists := deleteList[v.Data.ReleaseName]; exists {
					log.WithFields(log.Fields{
						"Name":    v.Metadata.Name,
						"Release": rel.ReleaseName,
					}).Info("deleting release")
					if err := session.DeleteRelease(rel); err != nil {
						return errors.Wrap(err, "error deleting list")
					}
				}
			}
		}
	}
	return nil
}
