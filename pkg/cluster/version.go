package cluster

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"k8s.io/helm/pkg/timeconv"

	"github.com/charter-oss/barrelman/pkg/cluster/driver"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type Versioner interface {
	ListManifests() ([]*Version, error)
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (s *Session) NewConfigMaps() *driver.ConfigMaps {
	return driver.NewConfigMaps(s.Clientset.CoreV1().ConfigMaps(s.Tunnel.Namespace))
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

	resourceName := fmt.Sprintf("%s.v%d.%s", versions.Name, version, NewSHA1Hash())

	log.WithFields(log.Fields{
		"Name": resourceName,
	}).Debug("creating rollback resource")

	if err := cmap.Create(resourceName, rls); err != nil {
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
	cmap := driver.NewConfigMaps(s.Clientset.CoreV1().ConfigMaps(s.Tunnel.Namespace))
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

// ListManifests returns list of unique Barrelman manifests recorded in cluster
func (s *Session) ListManifests() ([]*Version, error) {
	outVersions := []*Version{}
	internalData := make(map[string]map[string][]*Version)
	cmap := driver.NewConfigMaps(s.Clientset.CoreV1().ConfigMaps(s.Tunnel.Namespace))
	allReleases, err := cmap.List(getNoopManifestFilter())
	if err != nil {
		return nil, errors.Wrap(err, "ListManifests failed to get release list")
	}

	// order manifest revisions by namespace, then name
	for _, v := range allReleases {
		if _, ok := internalData[v.Namespace]; !ok {
			internalData[v.Namespace] = make(map[string][]*Version)
		}
		if _, ok := internalData[v.Namespace][v.Name]; !ok {
			internalData[v.Namespace][v.Name] = []*Version{}
		}

		internalData[v.Namespace][v.Name] = append(internalData[v.Namespace][v.Name], &Version{
			Name:      v.Name,
			Namespace: v.Namespace,
			Revision:  v.Version,
		})
	}

	// for each manifest, under a namespace, get the latest version
	for namespace, namespaceData := range internalData {
		for name, versions := range namespaceData {
			highest := int32(0)
			for _, v := range versions {
				if v.Revision > highest {
					highest = v.Revision
				}
			}
			outVersions = append(outVersions, &Version{
				Name:      name,
				Namespace: namespace,
				Revision:  highest,
			})
		}
	}

	return outVersions, nil
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

func (version *Version) ShortReport() map[string]interface{} {
	return map[string]interface{}{
		"Name":      version.Name,
		"Namespace": version.Namespace,
		"Revision":  version.Revision,
	}
}

func (version *Version) DetailedReport() map[string]interface{} {
	return map[string]interface{}{
		"Name":             version.Name,
		"Namespace":        version.Namespace,
		"Revision":         version.Revision,
		"PreviousRevision": version.PreviousRevision,
	}
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

func getNoopManifestFilter() func(rls *release.Release) bool {
	return func(rls *release.Release) bool {
		return true
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

func NewSHA1Hash(n ...int) string {
	noRandomCharacters := 32
	if len(n) > 0 {
		noRandomCharacters = n[0]
	}
	randString := RandomString(noRandomCharacters)
	hash := sha1.New()
	hash.Write([]byte(randString))
	bs := hash.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

var characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString generates a random string of n length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characterRunes[rand.Intn(len(characterRunes))]
	}
	return string(b)
}
