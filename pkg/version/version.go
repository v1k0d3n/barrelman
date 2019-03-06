package version

var (
	version = "UNSET"
	branch  = "UNSET"
	commit  = "UNSET"
)

type Version struct {
	Version string
	Branch  string
	Commit  string
}

func Get() *Version {
	return &Version{
		Version: version,
		Branch:  branch,
		Commit:  commit,
	}
}
