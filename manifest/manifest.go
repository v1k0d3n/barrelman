package manifest

import (
	"fmt"
	"strings"
)

const (
	Stringv1         = "v1"
	StringChartGroup = "ChartGroup"
	StringManifest   = "Manifest"
	StringChart      = "Chart"
)

type Schema struct {
	Route   string // armada, yaml2vars, etc
	Type    string // Openstack Deckhand document format spec has Document, and Control
	Version string
}

type LookupTable struct {
	Chart      map[string]*Chart
	ChartGroup map[string]*ChartGroup
}

type Manifest struct {
	Version string
	Name    string
	Data    *ManifestData
	Lookup  *LookupTable
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
	Version string
	Name    string
	Data    *ChartData
}

type ChartData struct {
	ChartName    string
	TestEnabled  bool
	Release      string
	Namespace    string
	Timeout      int
	Wait         *ChartDataWait
	Install      *ChartDataInstall
	Upgrade      *ChartDataUpgrade
	SubPath      string
	Location     string
	Values       map[string]string
	Dependencies []string
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

func NewManifest() *Manifest {
	manifest := &Manifest{}
	manifest.Data = &ManifestData{}
	manifest.Lookup = &LookupTable{}
	manifest.Lookup.Chart = make(map[string]*Chart)
	manifest.Lookup.ChartGroup = make(map[string]*ChartGroup)
	return manifest
}

func (m *Manifest) AddChartGroup(cg *ChartGroup) error {
	if _, exists := m.Lookup.ChartGroup[cg.Name]; exists {
		return fmt.Errorf("ChartGroup name already exists: %v", cg.Name)
	}
	m.Lookup.ChartGroup[cg.Name] = cg
	return nil
}

func (m *Manifest) GetChartGroup(s string) *ChartGroup {
	cg, _ := m.Lookup.ChartGroup[s]
	return cg
}

func (m *Manifest) AddChart(c *Chart) error {
	if _, exists := m.Lookup.Chart[c.Name]; exists {
		return fmt.Errorf("Chart name already exists: %v", c.Name)
	}
	m.Lookup.Chart[c.Name] = c
	fmt.Printf("Added chart %v\n", c.Name)
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
			return nil, fmt.Errorf("ChartGroup [%v] does not exist", name)
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
			return nil, fmt.Errorf("Chart [%v] does not exist")
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
	return chart
}

func ParseSchema(input string) (*Schema, error) {
	split := strings.Split(input, "/")
	if len(split) != 3 {
		return &Schema{}, fmt.Errorf("ParseSchema arrived at wrong number of elements from input %v", input)
	}
	schema := &Schema{
		Route:   split[0],
		Type:    split[1],
		Version: split[2],
	}
	return schema, nil
}
