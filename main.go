package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/log"
	"github.com/charter-se/barrelman/manifest"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	fmt.Printf("Barrelman Engage!\n")
	datadir := fmt.Sprintf("%v/.barrelman/data", userHomeDir())
	configFile := fmt.Sprintf("%v/.barrelman/config", userHomeDir())
	config, err := GetConfig(configFile)
	if err != nil {
		fmt.Printf("Got error while loading config: %v\n", err)
		os.Exit(1)
	}

	if err := ensureWorkDir(datadir); err != nil {
		fmt.Printf("Failed to create working directory: %v", err)
	}

	// Open connections to the k8s APIs
	c, err := cluster.NewSession(fmt.Sprintf("%v/.kube/config", userHomeDir()))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create new cluster session")
		return
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{DataDir: datadir})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while initializing manifest")
		return
	}

	log.Info("Syncronizing with remote chart repositories")
	if err := mfest.Sync(config.Account); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while downloading charts")
		return
	}

	if err := DeleteByManifest(mfest, c); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to delete by manifest")
		return
	}

	archives, err := mfest.CreateArchives()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create archives")
		return
	}

	for _, v := range archives.List {
		//Install the release from the tgz above
		relName, err := c.InstallRelease(&cluster.ReleaseMeta{
			Path:      v.Path,
			Namespace: v.Namespace,
		}, []byte{})
		if err != nil {
			fmt.Printf("Got ERROR: %v\n", err)
			return
		}
		log.WithFields(log.Fields{
			"Name":      v.Name,
			"Namespace": v.Namespace,
			"Release":   relName,
		}).Info("Installed release")
	}
	if err := archives.Purge(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to purge local archives")
		return
	}
}

func DeleteByManifest(bm *manifest.Manifest, c *cluster.Session) error {
	deleteList := make(map[string]*cluster.DeleteMeta)
	groups, err := bm.GetChartGroups()
	if err != nil {
		return fmt.Errorf("Error resolving chart groups: %v\n", err)
	}

	releases, err := c.ListReleases()
	if err != nil {
		return fmt.Errorf("Failed to list releases %v", err)
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
			return fmt.Errorf("Error resolving charts: %v\n", err)
		}
		for _, v := range charts {
			if dm, exists := deleteList[v.Name]; exists {

				log.WithFields(log.Fields{
					"Name":    v.Name,
					"Release": dm.Name,
				}).Info("Deleting release")
				if err := c.DeleteRelease(dm); err != nil {
					return fmt.Errorf("error deleting list: %v\n", err)
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
