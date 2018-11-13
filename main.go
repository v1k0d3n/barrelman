package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

	// Open connections to the k8s APIs
	c, err := cluster.NewSession(fmt.Sprintf("%v/.kube/config", userHomeDir()))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create new cluster session")
		return
	}

	// Open and initialize the manifest
	mfest := manifest.NewManifest()
	if err := mfest.Init(&manifest.Config{DataDir: datadir}); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while initializing manifest: %v", err)
		return
	}

	if err := ensureWorkDir(datadir); err != nil {
		fmt.Printf("Failed to create working directory: %v", err)
	}

	log.Info("Syncronizing with remote chart repositories")
	//Perform the chart syncronization/download/update whatever
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
	}

	//Chart groups as defined by Armada YAML spec
	groups, err := mfest.GetChartGroups()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error resolving chart groups: %v\n", err))
	}

	for _, cg := range groups {
		//All charts within the group
		charts, err := mfest.GetChartsByName(cg.Data.ChartGroup)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error resolving charts: %v\n", err))
			return
		}
		//For each chart within the group
		for _, chart := range charts {
			//Get the local file path and dependancies for the chart
			path, dependCharts, err := mfest.GetChartSpec(chart)
			if err != nil {
				log.Errorf("Error getting chart path: %v\n", err)
				return
			}
			//Build the tgz archive
			tgz, err := createChartArchive(datadir, path, dependCharts)
			if err != nil {
				log.WithFields(log.Fields{
					"archive": tgz,
					"error":   err,
				}).Error("failed to create tgz archive")
			}
			//Install the release from the tgz above
			relName, err := c.InstallRelease(&cluster.ReleaseMeta{
				Path:      tgz,
				Namespace: chart.Data.Namespace,
			}, []byte{})
			if err != nil {
				fmt.Printf("Got ERROR: %v\n", err)
				return
			}
			log.WithFields(log.Fields{
				"Name":      chart.Name,
				"Namespace": chart.Data.Namespace,
				"Release":   relName,
			}).Info("Installed release")
		}
	}
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
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

func createChartArchive(datadir string, path string, dependCharts []*manifest.ChartSpec) (string, error) {
	randomName := fmt.Sprintf("%v/%v", datadir, tempFileName("tmp_", ".tgz"))
	f, err := os.Create(randomName)
	if err != nil {
		return randomName, err
	}
	defer func() {
		f.Close()
	}()

	err = Package(dependCharts, path, f)
	return randomName, err
}

func ensureWorkDir(datadir string) error {
	return os.MkdirAll(datadir, os.ModePerm)
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return prefix + hex.EncodeToString(randBytes) + suffix
}

func Package(depends []*manifest.ChartSpec, src string, writers ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add main chart repo
	if err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		// update the name to correctly reflect the desired destination when untaring
		header.Name = fmt.Sprintf("this/%v", strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator)))
		if header.Name == "" {
			return err
		}
		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	}); err != nil {
		return err
	}

	//Add depends
	for _, v := range depends {
		if err := filepath.Walk(v.Path, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
			header.Name = fmt.Sprintf("this/charts/%v/%v", v.Name, strings.TrimPrefix(strings.Replace(file, v.Path, "", -1), string(filepath.Separator)))
			if header.Name == "" {
				return err
			}
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !fi.Mode().IsRegular() {
				return nil
			}

			f, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			f.Close()
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
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
