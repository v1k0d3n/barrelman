package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [manifest.yaml]",
	Short: "upgrade something",
	Long:  `Something something else...`,
	Args: func(cmd *cobra.Command, args []string) error {
		flagrx := regexp.MustCompile("^--")
		manifestFile = fmt.Sprintf("%v/.barrelman/manifest.yaml", userHomeDir())
		if len(args) > 0 {
			if args[0] != "" {
				if flagrx.FindAllStringSubmatchIndex(args[0], -1) == nil {
					manifestFile = args[0]
				}
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		runUpgrade(manifestFile)
	},
}

func runUpgrade(configFile string) {
	log.Warn("Barrelman Upgrade Engage!")
	configFile = fmt.Sprintf("%v/.barrelman/config", userHomeDir())
	datadir := fmt.Sprintf("%v/.barrelman/data", userHomeDir())
	config, err := GetConfig(configFile)
	if err != nil {
		log.Error(errors.Wrap(err, "got error while loading config"))
		os.Exit(1)
	}
	log.WithFields(log.Fields{"file": configFile}).Info("Using config")

	if err := ensureWorkDir(datadir); err != nil {
		log.Error(errors.Wrap(err, "failed to create working directory"))
		os.Exit(1)
	}

	// Open connections to the k8s APIs
	c, err := cluster.NewSession(fmt.Sprintf("%v/.kube/config", userHomeDir()))
	if err != nil {
		log.Error(errors.Wrap(err, "failed to create new cluster session"))
		os.Exit(1)
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{DataDir: datadir, ManifestFile: manifestFile})
	if err != nil {
		log.Error(errors.Wrap(err, "error while initializing manifest"))
		os.Exit(1)
	}

	log.Info("syncronizing with remote chart repositories")
	if err := mfest.Sync(config.Account); err != nil {
		log.Error(errors.Wrap(err, "error while downloading charts"))
		os.Exit(1)
	}

	archives, err := mfest.CreateArchives()
	if err != nil {
		log.Error(errors.Wrap(err, "failed to create archives"))
		os.Exit(1)
	}
	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()

	err = UpgradeByManifest(mfest, c)
	if err != nil {
		log.Error(errors.Wrap(err, "error while upgrading release"))
		return
	}
	// for _, v := range archives.List {
	// 	//Install the release from the tgz above
	// 	err := c.UpgradeRelease(&cluster.ReleaseMeta{
	// 		Path:      v.Path,
	// 		Namespace: v.Namespace,
	// 	})
	// 	if err != nil {
	// 		log.Error(errors.WithFields(errors.Fields{
	// 			"Name":      v.Name,
	// 			"Namespace": v.Namespace,
	// 		}).Wrap(err, "error while upgrading release"))
	// 		return
	// 	}
	// 	log.WithFields(log.Fields{
	// 		"Name":      v.Name,
	// 		"Namespace": v.Namespace,
	// 	}).Info("upgraded release")
	// }
}

func UpgradeByManifest(bm *manifest.Manifest, c *cluster.Session) error {
	upgradeList := make(map[string]*cluster.ReleaseMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return errors.Wrap(err, "error resolving chart groups")
	}

	releases, err := c.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		upgradeList[v.Chart.Metadata.Name] = &cluster.ReleaseMeta{
			Name:      v.Name,
			Namespace: v.Namespace,
		}
	}

	archives, err := bm.CreateArchives()
	if err != nil {
		log.Error(errors.Wrap(err, "failed to create archives"))
		os.Exit(1)
	}
	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()

	for _, v := range archives.List {
		//Install the release from the tgz above
		if rel, exists := upgradeList[v.Name]; exists {
			upgradeList[v.Name].Path = v.Path
			err := c.UpgradeRelease(&cluster.ReleaseMeta{
				Path:      v.Path,
				Name:      rel.Name,
				Namespace: v.Namespace,
			})
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.Name,
					"Namespace": v.Namespace,
					"Release":   rel.Name,
				}).Wrap(err, "error while installing release")
			}
			log.WithFields(log.Fields{
				"Name":      v.Name,
				"Path":      v.Path,
				"Namespace": v.Namespace,
				"Release":   rel.Name,
			}).Info("upgraded release")
		} else {
			relName, err := c.InstallRelease(&cluster.ReleaseMeta{
				Path:      v.Path,
				Namespace: v.Namespace,
			}, []byte{})
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Name":      v.Name,
					"Namespace": v.Namespace,
				}).Wrap(err, "error while installing release")
			}
			log.WithFields(log.Fields{
				"Name":      v.Name,
				"Namespace": v.Namespace,
				"Release":   relName,
			}).Info("installed release")
		}
	}

	for _, cg := range groups {
		charts, err := bm.GetChartsByName(cg.Data.ChartGroup)
		if err != nil {
			return errors.Wrap(err, "error resolving charts")
		}
		for _, v := range charts {
			if dm, exists := upgradeList[v.Name]; exists {
				log.WithFields(log.Fields{
					"Name":    v.Name,
					"Path":    dm.Path,
					"Release": dm.Name,
				}).Info("upgrading release")
				if err := c.UpgradeRelease(dm); err != nil {
					return errors.WithFields(errors.Fields{
						"Path": dm.Path,
					}).Wrap(err, "error upgrading list")
				}
			}
		}
	}
	return nil
}
