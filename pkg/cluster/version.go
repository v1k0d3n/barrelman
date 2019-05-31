package cluster

import (
	"fmt"
	"sort"
	"time"

	"k8s.io/helm/pkg/timeconv"

	"github.com/charter-oss/barrelman/pkg/cluster/driver"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type Versioner interface {
	GetVersionsFromList(manifestNames *[]string) ([]*Versions, error)
	GetVersions(manifestName string) (*Versions, error)
}

type Versions struct {
	Name string
	Data []*Version
}

type VersionTable struct {
	Name string
	Data map[int32]*Version
}

type VersionSync struct {
	Session *Session
}

type Version struct {
	Name             string
	Namespace        string
	Revision         int32
	PreviousRevision int32
	Chart            *chart.Chart
	Info             *release.Info
	Modified         bool
}

func (s *Session) NewConfigMaps() *driver.ConfigMaps {
	return driver.NewConfigMaps(s.Clientset.Core().ConfigMaps(s.Tiller.Namespace))
}

func (s *Session) WriteVersions(versions *Versions) error {
	cmap := s.NewConfigMaps()
	releases, err := cmap.List(getReleaseFilter(versions.Name))
	if err != nil {
		return errors.Wrap(err, "GetVersion failed to get release list during write")
	}

	rawReleaseValues, err := versions.RawReleaseTable()
	if err != nil {
		return err
	}
	releaseValues := versions.ChartValues()

	for r, v := range releaseValues {
		log.WithFields(log.Fields{
			"Release": r,
			"Version": v.GetValue(),
		}).Debug("Adding release to metaversion")
	}

	version := CalculateLastVersion(releases) + 1
	rls := &release.Release{
		Name: versions.Name,
		Info: &release.Info{LastDeployed: timeconv.Timestamp(time.Now())},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: versions.Name,
			},
			Values: &chart.Config{
				Raw:    rawReleaseValues,
				Values: releaseValues,
			},
		},
		Version: version,
	}

	log.Debug("creating rollback ConfigMap")
	if err := cmap.Create(fmt.Sprintf("%s.v%d", versions.Name, version), rls); err != nil {
		return err
	}

	return nil
}

func (s *Session) GetVersionsFromList(manifestNames *[]string) ([]*Versions, error) {
	allVersions := []*Versions{}
	for _, v := range *manifestNames {
		versions, err := s.GetVersions(v)
		if err != nil {
			return nil, errors.WithFields(errors.Fields{
				"ManifestName": v,
			}).Wrap(err, "failed to get current release versions for manifest")
		}
		allVersions = append(allVersions, versions)
	}
	return allVersions, nil
}

func (s *Session) GetVersions(manifestName string) (*Versions, error) {
	log.WithFields(log.Fields{
		"ManifestName": manifestName,
	}).Debug("Getting rollback information")
	versions := NewVersions(manifestName)
	cmap := driver.NewConfigMaps(s.Clientset.Core().ConfigMaps(s.Tiller.Namespace))
	releases, err := cmap.List(getReleaseFilter(manifestName))
	if err != nil {
		return nil, errors.Wrap(err, "GetVersion failed to get release list")
	}
	for _, v := range releases {
		versions.Data = append(versions.Data, &Version{
			Name:      v.Name,
			Namespace: v.Namespace,
			Revision:  v.Version,
			Chart:     v.Chart,
			Info:      v.Info,
		})
	}
	return versions, nil
}

// AddReleaseVersion add a release to the transaction for further processing
// Does not imply release has been modified
func (versions *Versions) AddReleaseVersion(rlsVersion *Version) error {
	log.WithFields(log.Fields{
		"Name":    rlsVersion.Name,
		"Version": rlsVersion.Revision,
	}).Debug("Add release version")
	versions.Data = append(versions.Data, rlsVersion)
	return nil
}

// Table returns a map of release group versions keyed by version
func (versions *Versions) Table() *VersionTable {
	versionTable := &VersionTable{
		Name: versions.Name,
		Data: make(map[int32]*Version),
	}
	for _, version := range versions.Data {
		log.WithFields(log.Fields{
			"ReleaseName": version.Name,
			"Revision":    version.Revision,
		}).Warn("Table() data")
		versionTable.Data[version.Revision] = version
	}
	return versionTable
}

func (versions *Versions) Lookup(name string) *Version {
	for _, version := range versions.Data {
		if version.Name == name {
			return version
		}
	}
	return nil
}

func (version *Version) ReleaseTable() (map[string]*chart.Value, error) {
	if version.Chart == nil {
		return nil, errors.New("Chart (ConfigMap) does not exist in version")
	}
	if version.Chart.Values == nil {
		return nil, errors.New("Values does not exist in version, cannot extract release table")
	}
	return version.Chart.Values.Values, nil
}

func (version *Version) SetRevision(newVersion int32) {
	version.Revision = newVersion
	version.SetModified()
}

func (version *Version) SetModified() {
	version.Modified = true
}

func (version *Version) IsModified() bool {
	return version.Modified
}

func (versions *Versions) ChartValues() map[string]*chart.Value {
	values := make(map[string]*chart.Value)
	for _, v := range versions.Data {
		log.WithFields(log.Fields{
			"ReleaseName": v.Name,
		}).Warn("ChartValues()")
		values[v.Name] = &chart.Value{fmt.Sprintf("%d", v.Revision)}
	}
	return values
}

func (versions *Versions) RawReleaseTable() (string, error) {
	type entry struct {
		Name      string
		Namespace string
		Revision  int32
	}
	type entries []*entry
	data := entries{}
	for _, v := range versions.Data {
		data = append(data, &entry{
			Name:      v.Name,
			Namespace: v.Namespace,
			Revision:  v.Revision,
		})
	}
	raw, err := yaml.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "Failed to marshal version information")
	}

	return string(raw), nil
}

func getReleaseFilter(manifestName string) func(rls *release.Release) bool {
	return func(rls *release.Release) bool {
		return rls.Name == manifestName
	}
}

func NewVersions(name string) *Versions {
	return &Versions{
		Name: name,
		Data: []*Version{},
	}
}

func CalculateLastVersion(releases []*release.Release) int32 {
	var highestVersion int32
	for _, rls := range releases {
		if rls.Version > highestVersion {
			highestVersion = rls.Version
		}
	}
	return highestVersion

}

// Version sorter for formatting of release information

// By is the type of a "less" function that defines the ordering of its arguments.
type By func(p1, p2 *Version) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(versions []*Version) {
	ps := &versionSorter{
		versions: versions,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// versionSorter joins a By function and a slice of Versions to be sorted.
type versionSorter struct {
	versions []*Version
	by       func(r1, r2 *Version) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (v *versionSorter) Len() int {
	return len(v.versions)
}

// Swap is part of sort.Interface.
func (v *versionSorter) Swap(i, j int) {
	v.versions[i], v.versions[j] = v.versions[j], v.versions[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (v *versionSorter) Less(i, j int) bool {
	return v.by(v.versions[i], v.versions[j])
}
