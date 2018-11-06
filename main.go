package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/charter-se/barrelman/chartsync"
	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/sourcetype"
	"github.com/charter-se/barrelman/yamlpack"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	fmt.Printf("Barrelman Engage!\n")
	datadir := fmt.Sprintf("%v/.barrelman/data", userHomeDir())

	// Open connections to the k8s APIs
	c, err := cluster.NewSession(fmt.Sprintf("%v/.kube/config", userHomeDir()))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create new cluster session")
		return
	}

	if err := ensureWorkDir(datadir); err != nil {
		fmt.Printf("Failed to create working directory: %v", err)
	}

	// list pods
	pods, err := c.Clientset.Core().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to retrieve pods")
		return
	}
	for _, p := range pods.Items {
		log.WithFields(log.Fields{
			"namespace": p.Namespace,
			"name":      p.Name,
		}).Info("Found pods")
	}

	yp := yamlpack.New()
	if err := yp.Import("testdata/armada-osh.yaml"); err != nil {
		fmt.Printf("Error importing \"this\": %v\n", err)
	}

	cs := chartsync.New()

	for name, f := range yp.Files {
		fmt.Println("_________________________")
		fmt.Printf("This: %v\n", name)
		for _, k := range f {
			fmt.Printf("Schema: %v\n", k.Viper.Get("schema"))
			fmt.Printf("Metdata Name: %v\n", k.Viper.Get("metadata.name"))
			fmt.Printf("Metdata Schema: %v\n", k.Viper.Get("metadata.schema"))
			// y, err := k.Yaml()
			// if err != nil {
			// 	fmt.Printf("Failed to marshal data: %v\n", err)
			// 	return
			// }

			//Add each chart to chartsync to download all missing charts
			typ, err := sourcetype.Parse(k.Viper.GetString("data.source.type"))
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"type":  typ,
					"name":  k.Viper.GetString("metadata.name"),
				}).Error("Failed to parse source type")
				return
			}
			cs.Add(&chartsync.ChartMeta{
				Name:       k.Viper.GetString("metadata.name"),
				Depends:    k.Viper.GetStringSlice("data.dependancies"),
				Groups:     k.Viper.GetStringSlice("data.chart_group"),
				SourceType: typ,
			})

			//fmt.Printf(">>>>\n%v\n<<<<\n", string(y))
			// f, err := writeTmpYaml(datadir, []byte(y))
			// if err != nil {
			// 	fmt.Printf("Failed to write yaml section to file: %v", err)
			// }
			// if err := c.InstallRelease(&cluster.ReleaseMeta{
			// 	Path:      f,
			// 	NameSpace: k.Viper.GetString("data.namespace"),
			// }, []byte(y)); err != nil {
			// 	fmt.Printf("Got ERROR: %v\n", err)
			// }

			//fmt.Printf("Data:\n%v\n", string(y))
		}
		fmt.Printf("\n\n")
	}

	if err := cs.Sync(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while downloading charts")
	}
	// fmt.Printf("Yaml Sections:\n")
	// for _, s := range yp.ListYamls() {
	// 	fmt.Printf("\t%v\n", s)
	// }

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

func writeTmpYaml(datadir string, yamlBytes []byte) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	n, err := gz.Write(yamlBytes)
	if err != nil {
		return "", fmt.Errorf("Error while compressing yaml: %v", err)
	}
	if n < len(b.Bytes()) {
		return "", fmt.Errorf("Failed to fully compress yaml")
	}
	gz.Close()
	randomName := fmt.Sprintf("%v/%v", datadir, tempFileName("tmp_", ".gz"))
	f, err := os.Create(randomName)
	if err != nil {
		return randomName, err
	}
	defer func() {
		f.Close()
	}()
	n, err = f.Write(b.Bytes())
	if n < len(b.Bytes()) {
		return randomName, fmt.Errorf("Failed to fully write yaml to file \"%v\", %v", randomName, err)
	}
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
