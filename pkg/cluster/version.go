package cluster

import (
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/charter-oss/barrelman/pkg/cluster/driver"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
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
	Name      string
	Namespace string
	Revision  int32
	Chart     *chart.Chart
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

	values, err := versions.RawReleaseTable()
	if err != nil {
		return err
	}

	version := CalculateLastVersion(releases) + 1
	rls := &release.Release{
		Name: versions.Name,
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: versions.Name,
			},
			Values: &chart.Config{
				Raw:    values,
				Values: versions.ChartValues(),
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
		log.WithFields(log.Fields{
			"ReleaseName": v.Name,
		}).Debug("Adding rollback release")
		versions.Data = append(versions.Data, &Version{
			Name:      v.Name,
			Namespace: v.Namespace,
			Revision:  v.Version,
			Chart:     v.Chart,
		})
	}
	return versions, nil
}

func (versions *Versions) AddReleaseVersion(rlsVersion *Version) error {
	log.WithFields(log.Fields{
		"Name":    rlsVersion.Name,
		"Version": rlsVersion.Revision,
	}).Debug("Add release version")
	versions.Data = append(versions.Data, rlsVersion)
	log.WithFields(log.Fields{
		"Len": len(versions.Data),
	}).Warn("version.Data")
	return nil
}

func (versions *Versions) Table() *VersionTable {
	versionTable := &VersionTable{
		Name: versions.Name,
		Data: make(map[int32]*Version),
	}
	for _, version := range versions.Data {
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

func (versions *Versions) ChartValues() map[string]*chart.Value {
	values := make(map[string]*chart.Value)
	for k, v := range versions.Table().Data {
		values[v.Name] = &chart.Value{fmt.Sprintf("%d", k)}
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
