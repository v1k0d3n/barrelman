package barrelman

type CmdOptions struct {
	ManifestFile   string
	ConfigFile     string
	KubeConfigFile string
	KubeContext    string
	DataDir        string
	DryRun         bool
	Diff           bool
	NoSync         bool
	Debug          bool
	Force          *[]string
}
