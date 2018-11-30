package cmd

import (
	"fmt"
	"os"
)

type Defaults struct {
	ManifestFile   string
	KubeConfigFile string
	KubeContext    string
	DataDir        string
	ConfigFile     string
	InstallRetry   int
	Force          *[]string
}

func Default() *Defaults {
	d := &Defaults{}
	d.ConfigFile = fmt.Sprintf("%v/.barrelman/config", userHomeDir())
	d.ManifestFile = fmt.Sprintf("%v/.barrelman/manifest.yaml", userHomeDir())
	d.KubeConfigFile = os.Getenv("KUBE_CONFIG")
	if d.KubeConfigFile == "" {
		d.KubeConfigFile = fmt.Sprintf("%v/.kube/config", userHomeDir())
	}
	d.KubeContext = os.Getenv("KUBE_CONTEXT")
	d.DataDir = fmt.Sprintf("%v/.barrelman/data", userHomeDir())
	d.InstallRetry = int(1)
	d.Force = &[]string{}
	return d
}
