package cmd

import "fmt"

type Defaults struct {
	ManifestFile   string
	KubeConfigFile string
	DataDir        string
	ConfigFile     string
}

func Default() *Defaults {
	d := &Defaults{}
	d.ConfigFile = fmt.Sprintf("%v/.barrelman/config", userHomeDir())
	d.ManifestFile = fmt.Sprintf("%v/.barrelman/manifest.yaml", userHomeDir())
	d.KubeConfigFile = fmt.Sprintf("%v/.kube/config", userHomeDir())
	d.DataDir = fmt.Sprintf("%v/.barrelman/data", userHomeDir())
	return d
}
