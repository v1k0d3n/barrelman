package manifest

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charter-se/barrelman/manifest/chartsync"
	"github.com/charter-se/barrelman/manifest/sourcetype"
	"github.com/charter-se/barrelman/manifest/yamlpack"
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
type Config struct {
	DataDir string
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

type ChartSpec struct {
	Name string
	Path string
}

type ArchiveSpec struct {
	Name      string
	Path      string
	Namespace string
}

type RemoteAccount struct {
	Type   string
	Name   string
	Secret string
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

func (m *Manifest) Init(c *Config) error {
	m.Config = c
	m.yp = yamlpack.New()
	if err := m.yp.Import("testdata/flagship-manifest.yaml"); err != nil {
		fmt.Printf("Error importing \"this\": %v\n", err)
	}
	m.ChartSync = chartsync.New(m.Config.DataDir)
	if err := m.load(); err != nil {
		return err
	}
	return nil
}

func (m *Manifest) Sync(config chartsync.AccountTable) error {
	for _, k := range m.yp.AllSections() {
		//Get the URI type in order for chartsync
		typ, err := sourcetype.Parse(k.GetString("data.source.type"))
		if err != nil {
			return fmt.Errorf("Failed to parse source type [%v]: %v", typ, err)
		}
		//Add each chart to repo to download/update all charts
		m.ChartSync.Add(&chartsync.ChartMeta{
			Name:       k.GetString("metadata.name"),
			Location:   k.GetString("data.source.location"),
			Depends:    k.GetStringSlice("data.dependancies"),
			Groups:     k.GetStringSlice("data.chart_group"),
			SourceType: typ,
		})
	}

	//log.Info("Syncronizing with chart repositories")
	//Perform the chart syncronization/download/update whatever

	if err := m.ChartSync.Sync(config); err != nil {
		return fmt.Errorf("Error while downloading charts: %v", err)
	}

	return nil
}

func (m *Manifest) load() error {
	for _, k := range m.yp.AllSections() {
		schem, err := parseSchema(k.GetString("schema"))
		if err != nil {
			return fmt.Errorf("Failed to parse schema %v: %v", k.GetString("metatdata.name"), err)
		}
		switch schem.Type {
		case StringChart:
			chart := NewChart()
			chart.Name = k.GetString("metadata.name")
			chart.Version = schem.Version
			chart.Data.ChartName = k.GetString("data.chart_name")
			chart.Data.Dependencies = k.GetStringSlice("data.dependencies")
			chart.Data.Namespace = k.GetString("data.namespace")
			chart.Data.SubPath = k.GetString("data.source.subpath")
			chart.Data.Location = k.GetString("data.source.location")
			if err := m.AddChart(chart); err != nil {
				return fmt.Errorf("Error loading chart: %v\n", err)
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
		return &Schema{}, fmt.Errorf("ParseSchema arrived at wrong number of elements from input %v", input)
	}
	schema := &Schema{
		Route:   split[0],
		Type:    split[1],
		Version: split[2],
	}
	return schema, nil
}

// func (m *Manifest) GetAllChartPaths() ([]*ChartSpec, error) {

// }

func (m *Manifest) GetChartSpec(c *Chart) (string, []*ChartSpec, error) {

	path, err := m.ChartSync.GetPath(&chartsync.ChartMeta{
		Name:     c.Name,
		Location: c.Data.Location,
		Depends:  c.Data.Dependencies,
		SubPath:  c.Data.SubPath,
	})
	if err != nil {
		return "", nil, fmt.Errorf("Failed to get yaml file path: %v", err)
	}
	dependCharts, err := func() ([]*ChartSpec, error) {
		ret := []*ChartSpec{}
		for _, v := range c.Data.Dependencies {
			thischart := m.GetChart(v)
			if thischart == nil {
				return nil, fmt.Errorf("Failed getting chart for %v", v)
			}
			thispath, err := m.ChartSync.GetPath(&chartsync.ChartMeta{
				Name:     thischart.Name,
				Location: thischart.Data.Location,
				Depends:  thischart.Data.Dependencies,
				SubPath:  thischart.Data.SubPath,
			})
			if err != nil {
				return nil, fmt.Errorf("Failed getting path")
			}
			ret = append(ret, &ChartSpec{Name: thischart.Name, Path: thispath})
		}
		return ret, nil
	}()
	return path, dependCharts, nil
}

func (m *Manifest) CreateArchives() ([]*ArchiveSpec, error) {
	archives := []*ArchiveSpec{}
	//Chart groups as defined by Armada YAML spec
	groups, err := m.GetChartGroups()
	if err != nil {
		return nil, fmt.Errorf("error resolving chart groups: %v", err)
	}

	for _, cg := range groups {
		//All charts within the group
		charts, err := m.GetChartsByName(cg.Data.ChartGroup)
		if err != nil {
			return nil, fmt.Errorf("error resolving charts: %v", err)
		}
		//For each chart within the group
		for _, chart := range charts {
			//Get the local file path and dependancies for the chart
			path, dependCharts, err := m.GetChartSpec(chart)
			if err != nil {
				return nil, fmt.Errorf("error getting chart path: %v", err)
			}
			//Build the tgz archive
			tgz, err := createChartArchive(m.Config.DataDir, path, dependCharts)
			if err != nil {
				return nil, fmt.Errorf("failed to create tgz archive %v: %v", tgz, err)
			}
			archives = append(archives, &ArchiveSpec{
				Name:      chart.Name,
				Path:      tgz,
				Namespace: chart.Data.Namespace,
			})
		}
	}
	return archives, nil
}

func createChartArchive(datadir string, path string, dependCharts []*ChartSpec) (string, error) {
	randomName := fmt.Sprintf("%v/%v", datadir, tempFileName("tmp_", ".tgz"))
	f, err := os.Create(randomName)
	if err != nil {
		return randomName, err
	}
	defer func() {
		f.Close()
	}()

	err = Package(dependCharts, path, f)
	return randomName, err
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return prefix + hex.EncodeToString(randBytes) + suffix
}

func Package(depends []*ChartSpec, src string, writers ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add main chart repo
	if err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		// update the name to correctly reflect the desired destination when untaring
		header.Name = fmt.Sprintf("this/%v", strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator)))
		if header.Name == "" {
			return err
		}
		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	}); err != nil {
		return err
	}

	//Add depends
	for _, v := range depends {
		if err := filepath.Walk(v.Path, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
			header.Name = fmt.Sprintf("this/charts/%v/%v", v.Name, strings.TrimPrefix(strings.Replace(file, v.Path, "", -1), string(filepath.Separator)))
			if header.Name == "" {
				return err
			}
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !fi.Mode().IsRegular() {
				return nil
			}

			f, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			f.Close()
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
