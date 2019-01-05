package manifest

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/charter-se/barrelman/manifest/chartsync"
	"github.com/charter-se/structured/errors"
	"github.com/cirrocloud/yamlpack"
)

const (
	Stringv1         = "v1"
	StringChartGroup = "ChartGroup"
	StringManifest   = "Manifest"
	StringChart      = "Chart"
)

type Schema struct {
	Route   string // armada, yaml2vars, etc
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
}
type Manifest struct {
	yp        *yamlpack.Yp
	ChartSync *chartsync.ChartSync
	Config    *Config
	Version   string
	Name      string
	Data      *ManifestData
	Lookup    *LookupTable
}

type ManifestData struct {
	ReleasePrefix string
	ChartGroups   []string
}

type ChartGroup struct {
	Version string
	Name    string
	Data    *ChartGroupData
}

type ChartGroupData struct {
	Description string
	Sequenced   bool
	ChartGroup  []string
}

type Chart struct {
	Version  string
	MetaName string
	Data     *ChartData
}

type ChartData struct {
	ChartName    string
	TestEnabled  bool
	ReleaseName  string
	Namespace    string
	Timeout      int
	Wait         *ChartDataWait
	Install      *ChartDataInstall
	Upgrade      *ChartDataUpgrade
	SubPath      string
	Type         string
	Source       *chartsync.Source
	Values       map[string]string
	Overrides    []byte
	Dependencies []string
	Archiver     chartsync.Archiver
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

//New creates an initializes a *Manifest instance
func New(c *Config) (*Manifest, error) {
	m := &Manifest{}
	m.Data = &ManifestData{}
	m.Lookup = &LookupTable{}
	m.Lookup.Chart = make(map[string]*Chart)
	m.Lookup.ChartGroup = make(map[string]*ChartGroup)

	if c.AccountTable == nil {
		return nil, errors.New("manifest.New() called without accoutn table")
	}
	m.Config = c
	m.yp = yamlpack.New()
	file := m.Config.ManifestFile
	fileR, err := os.Open(file)
	if err != nil {
		return &Manifest{}, errors.WithFields(errors.Fields{"file": file}).Wrap(err, "error opening file")
	}
	if err := m.yp.Import(file, fileR); err != nil {
		return &Manifest{}, errors.WithFields(errors.Fields{"file": file}).Wrap(err, "error importing manifest")
	}
	m.ChartSync = chartsync.New(m.Config.DataDir, c.AccountTable)
	if err := m.load(); err != nil {
		return nil, errors.Wrap(err, "Error running chartsync")
	}

	return m, nil
}

func (m *Manifest) AddChartGroup(cg *ChartGroup) error {
	if _, exists := m.Lookup.ChartGroup[cg.Name]; exists {
		return errors.WithFields(errors.Fields{"Name": cg.Name}).New("ChartGroup name already exists")
	}
	m.Lookup.ChartGroup[cg.Name] = cg
	return nil
}

func (m *Manifest) GetChartGroup(s string) *ChartGroup {
	cg, _ := m.Lookup.ChartGroup[s]
	return cg
}

func (m *Manifest) AddChart(c *Chart) error {
	if _, exists := m.Lookup.Chart[c.MetaName]; exists {
		return errors.WithFields(errors.Fields{"Name": c.MetaName}).New("Chart name already exists")
	}
	m.Lookup.Chart[c.MetaName] = c
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
			if iv.Data.ChartName == name {
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
	chart.Data.Values = make(map[string]string)
	chart.Data.Dependencies = []string{}
	chart.Data.Source = &chartsync.Source{}
	return chart
}

//Sync updates local copies of remote repositories configured in a manifest
func (m *Manifest) Sync() error {
	for _, c := range m.AllCharts() {
		//Add each chart to repo to download/update all charts
		if err := m.ChartSync.Add(&chartsync.ChartMeta{
			Name:    c.MetaName,
			Depends: c.Data.Dependencies,
			Type:    c.Data.Type,
			Source:  c.Data.Source,
		}); err != nil {
			return err
		}
	}

	if err := m.ChartSync.Sync(m.Config.AccountTable); err != nil {
		return errors.Wrap(err, "error while downloading charts")
	}
	return nil
}

func (m *Manifest) load() error {
	for _, k := range m.yp.AllSections() {
		schem, err := parseSchema(k.GetString("schema"))
		if err != nil {
			return errors.WithFields(errors.Fields{
				"Schema": k.GetString("metatdata.name"),
			}).Wrap(err, "Failed to parse schema")
		}
		switch schem.Type {
		case StringChart:
			chart := NewChart()
			chart.MetaName = k.GetString("metadata.name")
			chart.Version = schem.Version
			chart.Data.ChartName = k.GetString("data.chart_name")
			chart.Data.ReleaseName = k.GetString("data.release")
			chart.Data.Dependencies = k.GetStringSlice("data.dependencies")
			chart.Data.Namespace = k.GetString("data.namespace")
			chart.Data.Type = k.GetString("data.source.type")

			chart.Data.Overrides, err = yaml.Marshal(k.Viper.Sub("data.values").AllSettings())
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Type,
					"Name": chart.MetaName,
				}).Wrap(err, "Failed to marshal Override Values")
			}

			chart.Data.Source = &chartsync.Source{
				Location:  k.GetString("data.source.location"),
				SubPath:   k.GetString("data.source.subpath"),
				Reference: k.GetString("data.source.reference"),
			}

			handler, err := chartsync.GetHandler(chart.Data.Type)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Type,
					"Name": chart.MetaName,
				}).Wrap(err, "Failed to find handler for source")
			}
			chart.Data.Archiver, err = handler.New(m.Config.DataDir,
				&chartsync.ChartMeta{
					Name:    chart.MetaName,
					Source:  chart.Data.Source,
					Depends: chart.Data.Dependencies,
					Type:    chart.Data.Type,
				},
				m.Config.AccountTable,
			)
			if err != nil {
				return errors.WithFields(errors.Fields{
					"Type": chart.Data.Type,
					"Name": chart.MetaName,
				}).Wrap(err, "Failed to generate new handler")
			}

			if err := m.AddChart(chart); err != nil {
				return errors.Wrap(err, "Error loading chart")
			}

		case StringChartGroup:
			chartGroup := NewChartGroup()
			chartGroup.Name = k.GetString("metadata.name")
			chartGroup.Version = schem.Version
			chartGroup.Data.Description = k.GetString("data.description")
			chartGroup.Data.Sequenced = k.GetBool("data.sequenced")
			chartGroup.Data.ChartGroup = k.GetStringSlice("data.chart_group")
			m.AddChartGroup(chartGroup)

		case StringManifest:
			m.Name = k.GetString("metadata.name")
			m.Version = schem.Version
			m.Data.ChartGroups = k.GetStringSlice("data.chart_groups")
			m.Data.ReleasePrefix = k.GetString("data.release_prefix")
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
			ret = append(ret, &chartsync.ChartSpec{Name: dependchart.MetaName, Path: absPath})
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
	//Chart groups as defined by Armada YAML spec
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
			as, err := Archive(m.Config.DataDir, chart, path, dependCharts, chart.Data.Archiver)
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
