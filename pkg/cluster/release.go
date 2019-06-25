//go:generate mockery -name=Releaser
package cluster

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aryann/difflib"
	"github.com/mgutz/ansi"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

const (
	Status_DEPLOYED         = Status(release.Status_DEPLOYED)
	Status_FAILED           = Status(release.Status_FAILED)
	Status_PENDING_INSTALL  = Status(release.Status_PENDING_INSTALL)
	Status_PENDING_ROLLBACK = Status(release.Status_PENDING_ROLLBACK)
	Status_PENDING_UPGRADE  = Status(release.Status_PENDING_UPGRADE)
	Status_UNKNOWN          = Status(release.Status_UNKNOWN)
	Status_DELETED          = Status(release.Status_DELETED)
)

//ReleaseMeta is used with the InstallRelease, UpgradeRelease methods
type ReleaseMeta struct {
	Chart            *chart.Chart
	Config           *chart.Config
	Path             string //Location of Chartfile
	MetaName         string //As presented in manifest
	ChartName        string //Defined in manifest, but alignes in processing
	ReleaseName      string //Defined in manifest, or presented by k8s
	Namespace        string
	Status           Status
	Revision         int32
	ValueOverrides   []byte
	InstallDryRun    bool
	InstallReuseName bool
	InstallWait      bool
	InstallTimeout   time.Duration
	DryRun           bool
}

//DeleteMeta is used with the DeleteRelease method
type DeleteMeta struct {
	ReleaseName   string
	Namespace     string
	Purge         bool
	DeleteTimeout time.Duration
}

//RollbackMeta is used with the RollbackRelease method
type RollbackMeta struct {
	ReleaseName string
	Namespace   string
	Revision    int32
}

type Revision struct {
}

type Status release.Status_Code

type Chart = chart.Chart

//Release contains current release information
type Release struct {
	Chart       *Chart
	ReleaseName string
	Namespace   string
	Status      Status
	Revision    int32
	Config      *chart.Config
}

type InstallReleaseResponse struct {
	Description    string
	ReleaseName    string
	ReleaseVersion int32
}

type UpgradeReleaseResponse struct {
	Description    string
	ReleaseVersion int32
}

type ReleaseDiff struct {
}

type Releaser interface {
	ListReleases() ([]*Release, error)
	InstallRelease(m *ReleaseMeta, manifestName string) (*InstallReleaseResponse, error)
	DiffRelease(m *ReleaseMeta) (bool, []byte, error)
	UpgradeRelease(m *ReleaseMeta, manifestName string) (*UpgradeReleaseResponse, error)
	DeleteReleases(dm []*DeleteMeta) error
	DeleteRelease(m *DeleteMeta) error
	Releases() (map[string]*ReleaseMeta, error)
	ReleasesByManifest(manifest string) (map[string]*ReleaseMeta, error)
	DiffManifests(map[string]*MappingResult, map[string]*MappingResult, []string, int, io.Writer) bool
	ChartFromArchive(aChart io.Reader) (*chart.Chart, error)
	GetRelease(releaseName string, revision int32) (*ReleaseMeta, error)
	RollbackRelease(m *RollbackMeta) (int32, error)
}

//ListReleases returns an array of running releases as reported by the cluster
func (s *Session) ListReleases() ([]*Release, error) {
	return s.ListReleasesByManifest("")
}

//ListReleases by Manifest returns an array of running releases as reported by the cluster
// filtered by the ManifestName label
func (s *Session) ListReleasesByManifest(manifestName string) ([]*Release, error) {
	var releases []*Release
	var filteredReleases []*Release

	r, err := s.Helm.ListReleases(
		helm.ReleaseListStatuses([]release.Status_Code{
			release.Status_DELETED,
			release.Status_DEPLOYED,
			release.Status_FAILED,
			release.Status_PENDING_INSTALL,
			release.Status_PENDING_ROLLBACK,
			release.Status_PENDING_UPGRADE,
			release.Status_UNKNOWN,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Helm.ListReleases()")
	}
	for _, v := range r.GetReleases() {
		rel := &Release{
			Chart:       v.GetChart(),
			ReleaseName: v.Name,
			Namespace:   v.Namespace,
			Status:      Status(v.Info.Status.Code),
			Revision:    v.Version,
			Config:      v.Config,
		}
		releases = append(releases, rel)
	}
	if manifestName == "" {
		return releases, err
	}

	for _, v := range releases {
		if getChartManifestTag(v.Chart) == manifestName {
			filteredReleases = append(filteredReleases, v)
		}
	}
	return filteredReleases, err
}

//InstallRelease uploads a chart and starts a release
func (s *Session) InstallRelease(m *ReleaseMeta, manifestName string) (*InstallReleaseResponse, error) {
	res, err := s.Helm.InstallReleaseFromChart(
		setChartManifestTags(m.Chart, "Manifest="+manifestName),
		m.Namespace,
		helm.ReleaseName(m.ReleaseName),
		helm.ValueOverrides(m.ValueOverrides),
		helm.InstallDryRun(m.DryRun),
		helm.InstallReuseName(m.InstallReuseName),
		helm.InstallWait(m.InstallWait),
		helm.InstallTimeout(int64(m.InstallTimeout.Seconds())),
	)
	if err != nil {
		return &InstallReleaseResponse{}, errors.WithFields(errors.Fields{
			"File":      m.Path,
			"Name":      m.MetaName,
			"Namespace": m.Namespace,
		}).Wrap(err, "failed install")
	}

	return &InstallReleaseResponse{
		Description:    res.Release.Info.Description,
		ReleaseName:    res.Release.Name,
		ReleaseVersion: res.Release.Version,
	}, err
}

//DiffRelease compares the differences between a running release and a proposed release
func (s *Session) DiffRelease(m *ReleaseMeta) (bool, []byte, error) {
	buf := bytes.NewBufferString("")
	currentR, err := s.Helm.ReleaseContent(m.ReleaseName)
	if err != nil {
		return false, nil, errors.Wrap(err, "Upgrade failed to get current release")
	}
	currentParsed := ParseRelease(currentR.Release)
	res, err := s.Helm.UpdateReleaseFromChart(
		m.ReleaseName,
		m.Chart,
		helm.UpgradeDryRun(true),
		helm.UpdateValueOverrides(m.ValueOverrides),
	)
	if err != nil {
		return false, []byte{}, errors.Wrap(err, "Failed to get results from Tiller")
	}

	newParsed := ParseRelease(res.Release)

	manifestsChanged := DiffManifests(currentParsed, newParsed, []string{}, int(10), buf)
	valuesChanged := DiffOverrides(currentR.Release.Config.Raw, res.Release.Config.Raw, buf)
	return manifestsChanged || valuesChanged, buf.Bytes(), err
}

// GetRelease retrieves release data by release revision
func (s *Session) GetRelease(releaseName string, revision int32) (*ReleaseMeta, error) {
	currentR, err := s.Helm.ReleaseContent(releaseName, helm.ContentReleaseVersion(revision))
	if err != nil {
		return nil, errors.WithFields(errors.Fields{
			"ReleaseName": releaseName,
			"Revision":    revision,
		}).Wrap(err, "failed to get release by version")
	}
	return &ReleaseMeta{
		Chart:       currentR.Release.Chart,
		ReleaseName: currentR.Release.Name,
		Namespace:   currentR.Release.Namespace,
		Revision:    currentR.Release.Version,
		Config:      currentR.Release.Config,
	}, nil
}

//UpgradeRelease applies changes to an already running release, potentially triggering a restart
func (s *Session) UpgradeRelease(m *ReleaseMeta, manifestName string) (*UpgradeReleaseResponse, error) {
	res, err := s.Helm.UpdateReleaseFromChart(
		m.ReleaseName,
		setChartManifestTags(m.Chart, "Manifest="+manifestName),
		helm.UpgradeForce(true),
		helm.UpgradeDryRun(m.DryRun),
		helm.UpdateValueOverrides(m.ValueOverrides),
	)
	if err != nil {
		return &UpgradeReleaseResponse{}, errors.Wrap(err, "Error during UpgradeRelease")
	}
	return &UpgradeReleaseResponse{
		Description:    res.Release.Info.Description,
		ReleaseVersion: res.Release.Version,
	}, err
}

//DeleteReleases calls DeleteRelease on an array of Releases
func (s *Session) DeleteReleases(dm []*DeleteMeta) error {
	for _, v := range dm {
		if err := s.DeleteRelease(v); err != nil {
			return err
		}
	}
	return nil
}

//DeleteRelease runs a DeleteRelease command based on a release name
func (s *Session) DeleteRelease(m *DeleteMeta) error {
	_, err := s.Helm.DeleteRelease(
		m.ReleaseName,
		helm.DeletePurge(m.Purge),
		helm.DeleteTimeout(int64(m.DeleteTimeout.Seconds())),
	)
	if err != nil {
		return errors.New(grpc.ErrorDesc(err))
	}
	return nil
}

//RollbackRelease sets the deployed revision
func (s *Session) RollbackRelease(m *RollbackMeta) (int32, error) {
	resp, err := s.Helm.RollbackRelease(
		m.ReleaseName,
		helm.RollbackForce(true),
		helm.RollbackVersion(m.Revision),
	)
	if err != nil {
		return 0, errors.New(grpc.ErrorDesc(err))
	}

	return resp.Release.Version, nil
}

//Releases queries a cluster and returns a map of currently deployed releases
func (s *Session) Releases() (map[string]*ReleaseMeta, error) {
	return s.ReleasesByManifest("")
}

func (s *Session) ReleasesByManifest(manifestName string) (map[string]*ReleaseMeta, error) {
	ret := make(map[string]*ReleaseMeta)

	releaseList, err := s.ListReleasesByManifest(manifestName)
	if err != nil {
		return ret, errors.Wrap(err, "failed to list releases")
	}

	for _, v := range releaseList {
		ret[v.ReleaseName] = &ReleaseMeta{
			Chart:       v.Chart,
			ReleaseName: v.ReleaseName,
			ChartName:   v.Chart.GetMetadata().Name,
			Namespace:   v.Namespace,
			Status:      v.Status,
			Revision:    v.Revision,
			Config:      v.Config,
		}
	}
	return ret, nil
}

func setChartManifestTags(chart *Chart, tags string) *Chart {
	chart.Metadata.Tags = tags
	return chart
}

func getChartManifestTag(chart *Chart) string {
	rx := regexp.MustCompile(`Manifest=([\.|\-|\d|\w]+)`)
	tags := rx.FindStringSubmatch(chart.Metadata.Tags)
	if len(tags) > 0 {
		log.WithFields(log.Fields{
			"Manifest": tags[1],
		}).Debug("Got manifest Tag")
		return tags[1]
	}
	return ""
}

// The following was taken from https://github.com/databus23/helm-diff
// ********************************************************************
var yamlSeperator = []byte("\n---\n")

type MappingResult struct {
	Name    string
	Kind    string
	Content string
}

type metadata struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   struct {
		Namespace string
		Name      string
	}
}

func (m metadata) String() string {
	apiBase := m.ApiVersion
	sp := strings.Split(apiBase, "/")
	if len(sp) > 1 {
		apiBase = strings.Join(sp[:len(sp)-1], "/")
	}

	return fmt.Sprintf("%s, %s, %s (%s)", m.Metadata.Namespace, m.Metadata.Name, m.Kind, apiBase)
}

func scanYamlSpecs(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, yamlSeperator); i >= 0 {
		// We have a full newline-terminated line.
		return i + len(yamlSeperator), data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func splitSpec(token string) (string, string) {
	if i := strings.Index(token, "\n"); i >= 0 {
		return token[0:i], token[i+1:]
	}
	return "", ""
}

func DiffOverrides(current string, proposed string, to io.Writer) (changed bool) {
	if current != proposed {
		fmt.Fprintf(to, ansi.Color("Override Values has changed:", "magenta")+"\n")
		diffs := diffStrings(current, proposed)
		if len(diffs) > 0 {
			changed = true
		}
		printDiffRecords([]string{}, "values", 0, diffs, to)
	}
	return changed
}

func ParseRelease(release *release.Release) map[string]*MappingResult {
	manifest := release.Manifest
	for _, hook := range release.Hooks {
		manifest += "\n---\n"
		manifest += fmt.Sprintf("# Source: %s\n", hook.Path)
		manifest += hook.Manifest
	}
	return Parse(manifest, release.Namespace)
}

func Parse(manifest string, defaultNamespace string) map[string]*MappingResult {
	scanner := bufio.NewScanner(strings.NewReader(manifest))
	scanner.Split(scanYamlSpecs)
	//Allow for tokens (specs) up to 1M in size
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 1048576)
	//Discard the first result, we only care about everything after the first seperator
	scanner.Scan()

	result := make(map[string]*MappingResult)

	for scanner.Scan() {
		content := scanner.Text()
		if strings.TrimSpace(content) == "" {
			continue
		}
		var metadata metadata
		if err := yaml.Unmarshal([]byte(content), &metadata); err != nil {
			log.Error(errors.WithFields(errors.Fields{
				"Content": content,
			}).Wrap(err, "Can't unmarshal yaml"))
			os.Exit(1)
		}

		if metadata.Metadata.Namespace == "" {
			metadata.Metadata.Namespace = defaultNamespace
		}
		name := metadata.String()
		if _, ok := result[name]; !ok {
			result[name] = &MappingResult{
				Name:    name,
				Kind:    metadata.Kind,
				Content: content,
			}
		}
	}
	return result
}

func (s *Session) ChartFromArchive(aChart io.Reader) (*chart.Chart, error) {
	return ChartFromArchive(aChart)
}

func ChartFromArchive(aChart io.Reader) (*chart.Chart, error) {
	c, err := chartutil.LoadArchive(aChart)
	if err != nil {
		return nil, errors.Wrap(err, "chart from archive failed")
	}
	return c, nil
}

func (s *Session) DiffManifests(oldIndex, newIndex map[string]*MappingResult, suppressedKinds []string, context int, to io.Writer) bool {
	return DiffManifests(oldIndex, newIndex, suppressedKinds, context, to)
}
func DiffManifests(oldIndex, newIndex map[string]*MappingResult, suppressedKinds []string, context int, to io.Writer) bool {
	seenAnyChanges := false
	emptyMapping := &MappingResult{}
	for key, oldContent := range oldIndex {
		if newContent, ok := newIndex[key]; ok {
			if oldContent.Content != newContent.Content {
				// modified
				fmt.Fprintf(to, ansi.Color("%s has changed:", "yellow")+"\n", key)
				diffs := diffMappingResults(oldContent, newContent)
				if len(diffs) > 0 {
					seenAnyChanges = true
				}
				printDiffRecords(suppressedKinds, oldContent.Kind, context, diffs, to)
			}
		} else {
			// removed
			fmt.Fprintf(to, ansi.Color("%s has been removed:", "yellow")+"\n", key)
			diffs := diffMappingResults(oldContent, emptyMapping)
			if len(diffs) > 0 {
				seenAnyChanges = true
			}
			printDiffRecords(suppressedKinds, oldContent.Kind, context, diffs, to)
		}
	}

	for key, newContent := range newIndex {
		if _, ok := oldIndex[key]; !ok {
			// added
			fmt.Fprintf(to, ansi.Color("%s has been added:", "yellow")+"\n", key)
			diffs := diffMappingResults(emptyMapping, newContent)
			if len(diffs) > 0 {
				seenAnyChanges = true
			}
			printDiffRecords(suppressedKinds, newContent.Kind, context, diffs, to)
		}
	}
	return seenAnyChanges
}

func diffMappingResults(oldContent *MappingResult, newContent *MappingResult) []difflib.DiffRecord {
	return diffStrings(oldContent.Content, newContent.Content)
}

func diffStrings(before, after string) []difflib.DiffRecord {
	const sep = "\n"
	return difflib.Diff(strings.Split(before, sep), strings.Split(after, sep))
}

func printDiffRecords(suppressedKinds []string, kind string, context int, diffs []difflib.DiffRecord, to io.Writer) {
	for _, ckind := range suppressedKinds {
		if ckind == kind {
			str := fmt.Sprintf("+ Changes suppressed on sensitive content of type %s\n", kind)
			fmt.Fprintf(to, ansi.Color(str, "yellow"))
			return
		}
	}

	if context >= 0 {
		distances := calculateDistances(diffs)
		omitting := false
		for i, diff := range diffs {
			if distances[i] > context {
				if !omitting {
					fmt.Fprintln(to, "...")
					omitting = true
				}
			} else {
				omitting = false
				printDiffRecord(diff, to)
			}
		}
	} else {
		for _, diff := range diffs {
			printDiffRecord(diff, to)
		}
	}
}

func printDiffRecord(diff difflib.DiffRecord, to io.Writer) {
	text := diff.Payload

	switch diff.Delta {
	case difflib.RightOnly:
		fmt.Fprintf(to, "%s\n", ansi.Color("+ "+text, "green"))
	case difflib.LeftOnly:
		fmt.Fprintf(to, "%s\n", ansi.Color("- "+text, "red"))
	case difflib.Common:
		fmt.Fprintf(to, "%s\n", "  "+text)
	}
}

// Calculate distance of every diff-line to the closest change
func calculateDistances(diffs []difflib.DiffRecord) map[int]int {
	distances := map[int]int{}

	// Iterate forwards through diffs, set 'distance' based on closest 'change' before this line
	change := -1
	for i, diff := range diffs {
		if diff.Delta != difflib.Common {
			change = i
		}
		distance := math.MaxInt32
		if change != -1 {
			distance = i - change
		}
		distances[i] = distance
	}

	// Iterate backwards through diffs, reduce 'distance' based on closest 'change' after this line
	change = -1
	for i := len(diffs) - 1; i >= 0; i-- {
		diff := diffs[i]
		if diff.Delta != difflib.Common {
			change = i
		}
		if change != -1 {
			distance := change - i
			if distance < distances[i] {
				distances[i] = distance
			}
		}
	}

	return distances
}

// ********************************************************************
