package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/errors"
	"github.com/charter-se/barrelman/log"
	"github.com/charter-se/barrelman/manifest"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.Warn("Barrelman Engage!")
	datadir := fmt.Sprintf("%v/.barrelman/data", userHomeDir())
	configFile := fmt.Sprintf("%v/.barrelman/config", userHomeDir())
	config, err := GetConfig(configFile)
	if err != nil {
		log.Error(errors.Wrap(err, "got error while loading config")
		os.Exit(1)
	}

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
	mfest, err := manifest.New(&manifest.Config{DataDir: datadir})
	if err != nil {
		log.Error(errors.Wrap(err, "error while initializing manifest"))
		os.Exit(1)
	}

	log.Info("syncronizing with remote chart repositories")
	if err := mfest.Sync(config.Account); err != nil {
		log.Error(errors.Wrap(err, "error while downloading charts"))
		os.Exit(1)
	}

	if err := DeleteByManifest(mfest, c); err != nil {
		log.Error(errors.Wrap(err, "failed to delete by manifest"))
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

	for _, v := range archives.List {
		//Install the release from the tgz above
		relName, err := c.InstallRelease(&cluster.ReleaseMeta{
			Path:      v.Path,
			Namespace: v.Namespace,
		}, []byte{})
		if err != nil {
			log.Error(errors.Wrap(err, "error while installing release"))
			return
		}
		log.WithFields(log.Fields{
			"Name":      v.Name,
			"Namespace": v.Namespace,
			"Release":   relName,
		}).Info("installed release")
	}
}

func DeleteByManifest(bm *manifest.Manifest, c *cluster.Session) error {
	deleteList := make(map[string]*cluster.DeleteMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return errors.Wrap(err, "error resolving chart groups")
	}

	releases, err := c.ListReleases()
	if err != nil {
		return errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releases {
		deleteList[v.Chart.Metadata.Name] = &cluster.DeleteMeta{
			Name:      v.Name,
			Namespace: "",
		}
	}

	for _, cg := range groups {
		charts, err := bm.GetChartsByName(cg.Data.ChartGroup)
		if err != nil {
			return errors.Wrap(err, "error resolving charts")
		}
		for _, v := range charts {
			if dm, exists := deleteList[v.Name]; exists {
				log.WithFields(log.Fields{
					"Name":    v.Name,
					"Release": dm.Name,
				}).Info("deleting release")
				if err := c.DeleteRelease(dm); err != nil {
					return errors.Wrap(err, "error deleting list")
				}
			}
		}
	}
	return nil
}

func ensureWorkDir(datadir string) error {
	return os.MkdirAll(datadir, os.ModePerm)
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
