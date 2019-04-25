//go:generate
package manifest

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	mock_chartsync "github.com/charter-oss/barrelman/pkg/manifest/chartsync/mocks"
)

func TestNewManifest(t *testing.T) {
	Convey("Manifest", t, func() {
		Convey("New can create new manifest instance", func() {
			config := &Config{
				ManifestFile: getTestDataDir() + "/unit-test-manifest.yaml",
				AccountTable: make(chartsync.AccountTable),
			}
			_, err := New(config)
			So(err, ShouldBeNil)
		})
		Convey("New can fail to open file", func() {
			_, err := New(&Config{
				ManifestFile: "somefile",
				AccountTable: make(chartsync.AccountTable),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})
		Convey("New can fail to import", func() {
			_, err := New(&Config{
				ManifestFile: getTestDataDir() + "/non-yaml",
				AccountTable: make(chartsync.AccountTable),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Failed to unmarshal schema")
		})
		Convey("New can fail to chartsync", func() {
			_, err := New(&Config{
				ManifestFile: getTestDataDir() + "/config",
				AccountTable: make(chartsync.AccountTable),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unable to parse schema")
		})
	})
}

func TestManifest(t *testing.T) {
	m := &Manifest{
		Lookup: &LookupTable{
			ChartGroup: make(map[string]*ChartGroup),
			Chart:      make(map[string]*Chart),
		},
	}
	newCG := &ChartGroup{Metadata: &Metadata{Name: "storage-minio"}}

	Convey("AddChartGroup", t, func() {
		Convey("New can create new manifest instance", func() {
			err := m.AddChartGroup(newCG)
			So(err, ShouldBeNil)
			So(m.Lookup.ChartGroup, ShouldContainKey, "storage-minio")
		})
		Convey("New can fail to add second instance", func() {
			err := m.AddChartGroup(newCG)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "already exists")
		})
	})
	Convey("GetChartGroup", t, func() {
		Convey("Can get chartgroup", func() {
			cg := m.GetChartGroup("storage-minio")
			So(cg, ShouldNotBeNil)
		})
		Convey("Can fail to get chartgroup", func() {
			cg := m.GetChartGroup("not-exist")
			So(cg, ShouldBeNil)
		})
	})
	Convey("AddChart", t, func() {
		Convey("Can add", func() {
			err := m.AddChart(&Chart{
				Metadata: &Metadata{Name: "storage-minio"},
			})
			So(err, ShouldBeNil)
		})
		Convey("Can fail to add", func() {
			err := m.AddChart(&Chart{
				Metadata: &Metadata{Name: "storage-minio"},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "name already exists")
		})
	})
	Convey("GetChart", t, func() {
		Convey("Can succeed", func() {
			chart := m.GetChart("storage-minio")
			So(chart, ShouldNotBeNil)
		})
		Convey("Can fail", func() {
			chart := m.GetChart("not-exists")
			So(chart, ShouldBeNil)
		})
	})
	Convey("AllChartGroups", t, func() {
		Convey("Can succeed", func() {
			thisCG := m.AllChartGroups()
			So(thisCG, ShouldNotBeNil)
			So(thisCG, ShouldHaveLength, 1)
			So(thisCG[0].Metadata.Name, ShouldEqual, "storage-minio")
		})
	})
	Convey("GetChartGroups", t, func() {
		m.Data = &ManifestData{
			ChartGroups: []string{
				"storage-minio",
			},
		}
		Convey("Can succeed", func() {
			thisCG, err := m.GetChartGroups()
			So(err, ShouldBeNil)
			So(thisCG, ShouldHaveLength, 1)
		})
		Convey("Can fail", func() {
			m.Data = &ManifestData{
				ChartGroups: []string{
					"does-not-compute",
				},
			}
			thisCG, err := m.GetChartGroups()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "does-not-compute")
			So(thisCG, ShouldHaveLength, 0)
		})
	})
	Convey("GetChartsNyName", t, func() {
		m.Data = &ManifestData{
			ChartGroups: []string{
				"storage-minio",
			},
		}
		Convey("Can succeed", func() {
			charts, err := m.GetChartsByName([]string{"storage-minio"})
			So(err, ShouldBeNil)
			So(charts, ShouldHaveLength, 1)
			So(charts[0].Metadata.Name, ShouldEqual, "storage-minio")
		})
		Convey("Can fail", func() {
			charts, err := m.GetChartsByName([]string{"no-exist"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "does not exist")
			So(charts, ShouldHaveLength, 0)
		})
	})
	Convey("AllCharts", t, func() {
		m.Data = &ManifestData{
			ChartGroups: []string{
				"storage-minio",
			},
		}
		Convey("Can succeed", func() {
			charts := m.AllCharts()
			So(charts, ShouldHaveLength, 1)
			So(charts[0].Metadata.Name, ShouldEqual, "storage-minio")
		})
	})
	Convey("GetChartSpec", t, func() {
		archiver := &mock_chartsync.Archiver{}
		archiver.On("GetPath").Return("charts/test-minio", nil).Maybe()
		m.Data = &ManifestData{
			ChartGroups: []string{
				"storage-minio",
			},
		}
		Convey("Can succeed", func() {
			m.ChartSync = chartsync.New(getTestDataDir(), make(chartsync.AccountTable))
			path, charts, err := m.GetChartSpec(&Chart{
				Metadata: &Metadata{Name: "storage-minio"},
				Data: &ChartData{
					Archiver:     archiver,
					ChartName:    "storage-minio",
					Dependencies: []string{},
					Source: &ChartSource{
						Subpath: "test-minio",
						Type:    "git",
					},
					SyncSource: &chartsync.Source{
						Location: "charts",
					},
				},
			})

			So(err, ShouldBeNil)
			So(path, ShouldContainSubstring, "charts/test-minio")
			So(charts, ShouldHaveLength, 0) // no depends in this test
		})
		Convey("Can process dependencies", func() {
			m.ChartSync = chartsync.New(getTestDataDir(), make(chartsync.AccountTable))
			m.AddChart(&Chart{
				Metadata: &Metadata{Name: "test-chart"},
				Data: &ChartData{
					Archiver: archiver,
					Source:   &ChartSource{Type: "git"},
					SyncSource: &chartsync.Source{
						Location: "charts",
					},
				},
			})
			path, charts, err := m.GetChartSpec(&Chart{
				Metadata: &Metadata{Name: "storage-minio"},
				Data: &ChartData{
					Archiver: archiver,
					Source: &ChartSource{
						Type:    "git",
						Subpath: "test-minio",
					},
					ChartName: "storage-minio",
					SyncSource: &chartsync.Source{
						Location: "charts",
					},
					Dependencies: []string{"test-chart"},
				},
			})
			So(err, ShouldBeNil)
			So(path, ShouldContainSubstring, "charts/test-minio")
			So(charts, ShouldHaveLength, 1)
		})
	})
}

//getTestDataDir returns a string representing the location of the testdata dir ectory as derived from THIS source file
//our tests are run in temporary directories, so finding the testdata can be a little troublesome
func getTestDataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return fmt.Sprintf("%v/../../testdata", path.Dir(filename))
}
