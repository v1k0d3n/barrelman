package manifest

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/cirrocloud/yamlpack"
	"github.com/ghodss/yaml"

	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/cirrocloud/structured"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

const (
	Stringv1         = "v1"
	StringChartGroup = "ChartGroup"
	StringManifest   = "Manifest"
	StringChart      = "Chart"
)

type Schema struct {
	Route   string // barrelman, yaml2vars, etc
	Type    string // Openstack Deckhand document format spec has "Document", and "Control"
	Version string
}

type LookupTable struct {
	Chart      map[string]*Chart
	ChartGroup map[string]*ChartGroup
}

type Config struct {
	DataDir      string
	ManifestFile string
	AccountTable chartsync.AccountTable
	Log          structured.Logger
}

type Manifest struct {
	ChartSync *chartsync.ChartSync
	Config    *Config
	Version   string
	Name      string
	Data      *ManifestData
	Lookup    *LookupTable
	YamlSec   []*yamlpack.YamlSection
}

type ManifestData struct {
	ReleasePrefix string   `json:"release_prefix" yaml:"release_prefix"`
	ChartGroups   []string `json:"chart_groups" yaml:"chart_groups"`
}

type ChartGroup struct {
	Version  string
	Metadata *Metadata
	Data     *ChartGroupData
}

type ChartGroupData struct {
	Description string
	Sequenced   bool
	ChartGroup  []string `json:"chart_group" yaml:"chart_group"`
}

type Chart struct {
	Version  string
	Metadata *Metadata
	Data     *ChartData
}

type ChartData struct {
	Archiver     chartsync.Archiver
	SyncSource   *chartsync.Source
	TestEnabled  bool
	Overrides    []byte
	ChartName    string `json:"chart_name" yaml:"chart_name"`
	ReleaseName  string `json:"release" yaml:"release"`
	Namespace    string
	Timeout      int
	Wait         *ChartDataWait
	Install      *ChartDataInstall
	InstallWait  bool
	Upgrade      *ChartDataUpgrade
	Source       *ChartSource
	Dependencies []string
	Values       map[string]interface{}
}

type ChartSource struct {
	Type      string
	Location  string
	Subpath   string
	Reference string
}

type ChartDataWait struct {
	Timeout int
	Labels  map[string]string
}

type ChartDataInstall struct {
	NoHooks bool
}

type ChartDataUpgrade struct {
	NoHooks bool
	//Investigate usage in HELM API
}

type RemoteAccount struct {
	Type   string
	Name   string
	Secret string
}

type Metadata struct {
	Schema string
	Name   string
}

//New creates an initializes a *Manifest instance
func New(c *Config) (*Manifest, error) {
	if c.Log == nil {
		c.Log = log.New()
	}
	file := c.ManifestFile
	c.Log.WithFields(log.Fields{
		"File": file,
	}).Debug("opening file")
	fileR, err := os.Open(file)
	if err != nil {
		return &Manifest{}, errors.WithFields(errors.Fields{"file": file}).Wrap(err, "error opening file")
	}
	defer fileR.Close()
	yp := yamlpack.New()
	if err := yp.Import(file, fileR); err != nil {
		return nil, errors.WithFields(errors.Fields{"file": file}).Wrap(err, "error importing manifest")
	}
	return NewFromSections(c, yp.AllSections())
}

//NewFromSections takes a *Config and array of *yamlpack.Section to assemble a *Manifest
func NewFromSections(c *Config, yamlSections []*yamlpack.YamlSection) (*Manifest, error) {
	m := &Manifest{}
	m.Data = &ManifestData{}
	m.Lookup = &LookupTable{}
	m.Lookup.Chart = make(map[string]*Chart)
	m.Lookup.ChartGroup = make(map[string]*ChartGroup)
	if c.Log == nil {
		c.Log = log.New()
	}
	if c.AccountTable == nil {
		return nil, errors.New("manifest.New() called without account table")
	}
	m.Config = c

	m.YamlSec = yamlSections
	m.ChartSync = chartsync.New(m.Config.DataDir, c.AccountTable)
	if err := m.load(); err != nil {
		return nil, errors.Wrap(err, "Error running chartsync")
	}
	return m, nil
}

func (m *Manifest) AddChartGroup(cg *ChartGroup) error {
	if _, exists := m.Lookup.ChartGroup[cg.Metadata.Name]; exists {
		return errors.WithFields(errors.Fields{"Name": cg.Metadata.Name}).New("ChartGroup name already exists")
	}
	m.Lookup.ChartGroup[cg.Metadata.Name] = cg
	return nil
}

func (m *Manifest) GetChartGroup(s string) *ChartGroup {
	cg, _ := m.Lookup.ChartGroup[s]
	return cg
}

func (m *Manifest) AddChart(c *Chart) error {
	if _, exists := m.Lookup.Chart[c.Metadata.Name]; exists {
		return errors.WithFields(errors.Fields{"Name": c.Metadata.Name}).New("Chart name already exists")
	}
	m.Lookup.Chart[c.Metadata.Name] = c
	return nil
}

func (m *Manifest) GetChart(s string) *Chart {
	c, _ := m.Lookup.Chart[s]
	return c
}

func (m *Manifest) AllChartGroups() []*ChartGroup {
	ret := []*ChartGroup{}
	for _, v := range m.Lookup.ChartGroup {
		ret = append(ret, v)
	}
	return ret
}

func (m *Manifest) GetChartGroups() ([]*ChartGroup, error) {
	ret := []*ChartGroup{}
	for _, name := range m.Data.ChartGroups {
		if v, exists := m.Lookup.ChartGroup[name]; exists {
			ret = append(ret, v)
		} else {
			return nil, errors.WithFields(errors.Fields{"Name": name}).New("ChartGroup does not exist")
		}
	}
	return ret, nil
}

func (m *Manifest) GetChartsByChartName(charts []string) ([]*Chart, error) {
	ret := []*Chart{}
	for _, name := range charts {
		chartExists := false
		for _, iv := range m.Lookup.Chart {
			if iv.Metadata.Name == name {
				chartExists = true
				ret = append(ret, iv)
			}
		}
		if !chartExists {
			return nil, errors.WithFields(errors.Fields{"Name": name}).New("Chart does not exist")
		}
	}
	return ret, nil
}

func (m *Manifest) GetChartsByName(charts []string) ([]*Chart, error) {
	ret := []*Chart{}
	for _, name := range charts {
		if v, exists := m.Lookup.Chart[name]; exists {
			ret = append(ret, v)
		} else {
			return nil, errors.WithFields(errors.Fields{"Name": name}).New("Chart does not exist")
		}
	}
	return ret, nil
}

func NewChartGroup() *ChartGroup {
	chartGroup := &ChartGroup{}
	chartGroup.Data = &ChartGroupData{}
	return chartGroup
}

func (m *Manifest) AllCharts() []*Chart {
	ret := []*Chart{}
	for _, v := range m.Lookup.Chart {
		ret = append(ret, v)
	}
	return ret
}

func NewChart() *Chart {
	chart := &Chart{}
	chart.Data = &ChartData{}
	chart.Data.Wait = &ChartDataWait{}
	chart.Data.Wait.Labels = make(map[string]string)
	chart.Data.Install = &ChartDataInstall{}
	chart.Data.Upgrade = &ChartDataUpgrade{}
	chart.Data.Values = make(map[string]interface{})
	chart.Data.Dependencies = []string{}
	chart.Data.SyncSource = &chartsync.Source{}
	return chart
}

//Sync updates local copies of remote repositories configured in a manifest
func (m *Manifest) Sync() error {
	for _, c := range m.AllCharts() {
		//Add each chart to repo to download/update all charts
		if err := m.ChartSync.Add(&chartsync.ChartMeta{
			Name:    c.Metadata.Name,
			Depends: c.Data.Dependencies,
			Type:    c.Data.Source.Type,
			Source:  c.Data.SyncSource,
		}); err != nil {
			return errors.WithFields(errors.Fields{
				"Name":   c.Metadata.Name,
				"Type":   c.Data.Source.Type,
				"Source": c.Data.SyncSource,
			}).Wrap(err, "Sync() failed")
		}
	}

	if err := m.ChartSync.Sync(m.Config.AccountTable); err != nil {
		return errors.Wrap(err, "error while downloading charts")
	}
	return nil
}

func (m *Manifest) load() error {
	for _, k := range m.YamlSec {
		schem, err := parseSchema(k.GetString("schema"))
		if err != nil {
			return errors.WithFields(errors.Fields{
				"Schema": k.GetString("metatdata.name"),
			}).Wrap(err, "Failed to parse schema")
		}
		switch schem.Type {
		case StringManifest:
			m.Version = schem.Version
			err := yaml.Unmarshal(k.Bytes, m)
			if err != nil {
				return errors.Wrap(err, "Error loading manifest")
			}
			m.Name = k.GetString("metadata.name")
		case StringChartGroup:
			chartGroup := NewChartGroup()
			chartGroup.Version = schem.Version
			err := yaml.Unmarshal(k.Bytes, &chartGroup)
			if err != nil {
				return err
			}
			if err := m.AddChartGroup(chartGroup); err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chartGroup.Metadata.Name,
					"Name": chartGroup.Metadata.Name,
				}).Wrap(err, "Failed to marshal Override Values")
			}
		case StringChart:
			chart := NewChart()
			chart.Version = schem.Version
			err := yaml.Unmarshal(k.Bytes, &chart)
			if err != nil {
				return err
			}

			chart.Data.Overrides, err = yaml.Marshal(chart.Data.Values)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Source.Type,
					"Name": chart.Metadata.Name,
				}).Wrap(err, "Failed to marshal Override Values")
			}
			if chart.Data.Source == nil {
				return errors.New("chart missing Data.Source")
			}
			chart.Data.SyncSource = &chartsync.Source{
				Location:  chart.Data.Source.Location,
				SubPath:   chart.Data.Source.Subpath,
				Reference: chart.Data.Source.Reference,
			}

			handler, err := chartsync.GetHandler(chart.Data.Source.Type)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Source.Type,
					"Name": chart.Metadata.Name,
				}).Wrap(err, "Failed to find handler for source")
			}
			chart.Data.Archiver, err = handler.New(
				m.Config.DataDir,
				&chartsync.ChartMeta{
					Name:    chart.Metadata.Name,
					Source:  chart.Data.SyncSource,
					Depends: chart.Data.Dependencies,
					Type:    chart.Data.Source.Type,
				},
				m.Config.AccountTable,
			)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Source.Type,
					"Name": chart.Metadata.Name,
				}).Wrap(err, "Failed to generate new handler")
			}

			if err := m.AddChart(chart); err != nil {
				return errors.Wrap(err, "Error loading chart")
			}
		}
	}
	return nil
}

func parseSchema(input string) (*Schema, error) {
	split := strings.Split(input, "/")
	if len(split) != 3 {
		return &Schema{}, errors.WithFields(errors.Fields{"Input": input}).New("ParseSchema arrived at wrong number of elements from input")
	}
	schema := &Schema{
		Route:   split[0],
		Type:    split[1],
		Version: split[2],
	}
	return schema, nil
}

//GetChartSpec returns local chart path and dependancy information useful for building an archive
func (m *Manifest) GetChartSpec(c *Chart) (string, []*chartsync.ChartSpec, error) {
	path, err := c.Data.Archiver.GetPath()
	if err != nil {
		return "", nil, errors.Wrap(err, "Failed to get yaml file path")
	}

	dependCharts, err := func() ([]*chartsync.ChartSpec, error) {
		ret := []*chartsync.ChartSpec{}
		for _, v := range c.Data.Dependencies {
			dependchart := m.GetChart(v)
			if dependchart == nil {
				return nil, errors.WithFields(errors.Fields{
					"Dependancy": v,
				}).New("failed getting depended chart")
			}
			dependpath, err := dependchart.Data.Archiver.GetPath()
			if err != nil {
				return nil, errors.Wrap(err, "Failed getting path")
			}
			absPath, err := filepath.Abs(dependpath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get absolute path")
			}
			ret = append(ret, &chartsync.ChartSpec{Name: dependchart.Metadata.Name, Path: absPath})
		}
		return ret, nil
	}()
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to compute dependencies")
	}
	return path, dependCharts, nil
}

//CreateArchives creates archives for charts configured in the manifest
func (m *Manifest) CreateArchives() (*ArchiveFiles, error) {
	af := &ArchiveFiles{List: []*ArchiveSpec{}}
	//Chart groups as defined by Barrelman YAML spec
	groups, err := m.GetChartGroups()
	if err != nil {
		return nil, errors.Wrap(err, "error resolving chart groups")
	}

	for _, cg := range groups {
		//All charts within the group
		charts, err := m.GetChartsByChartName(cg.Data.ChartGroup)
		if err != nil {
			return nil, errors.Wrap(err, "error resolving charts")
		}
		//For each chart within the group
		for _, chart := range charts {
			path, dependCharts, err := m.GetChartSpec(chart)
			if err != nil {
				return nil, errors.Wrap(err, "error getting chart path")
			}
			as, err := Archive(chart, path, dependCharts, chart.Data.Archiver)
			if err != nil {
				return nil, errors.Wrap(err, "Got err while running Archive")
			}
			af.List = append(af.List, as)
		}
	}
	return af, nil
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return prefix + hex.EncodeToString(randBytes) + suffix
}
