package cluster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

func TestVersioning(t *testing.T) {
	manifestName := "testManifest"
	s := NewMockSession()
	s.Tunnel.Namespace = "system"
	Convey("GetVersions", t, func() {
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)

		// Items will need to be populated with pkg/cluster/driver/util.go:decodeRelease()
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil)

		versions, err := s.GetVersions(manifestName)
		So(err, ShouldBeNil)
		So(versions.Data, ShouldHaveLength, 0)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)

		Convey("versions tooling", func() {
			newReleaseName := "newRelease"
			newRevision := int32(3)
			newRelease := &Version{
				Name:             newReleaseName,
				Namespace:        "oltherNamespace",
				Revision:         newRevision,
				PreviousRevision: int32(2),
				Chart:            &chart.Chart{},
				Info:             &release.Info{},
				Modified:         true,
			}
			err := versions.AddReleaseVersion(newRelease)
			So(err, ShouldBeNil)
			So(versions.Data, ShouldHaveLength, 1)
			Convey("Table should contain newRelease", func() {
				versionTable := versions.Table()
				So(versionTable.Data, ShouldHaveLength, 1)
				So(versionTable.Data, ShouldContainKey, newRevision)
			})
			Convey("Lookup can fail", func() {
				So(versions.Lookup("noFind"), ShouldBeNil)
			})
			Convey("Lookup can succeed", func() {
				ver := versions.Lookup(newReleaseName)
				So(ver.Name, ShouldEqual, newReleaseName)
				So(ver.Revision, ShouldEqual, newRevision)
				Convey("release.SetRevision sets revision", func() {
					nexRevision := newRevision + 1
					ver.SetRevision(nexRevision)
					So(ver.Revision, ShouldEqual, nexRevision)
				})
				Convey("SetModified sets modified", func() {
					ver.Modified = false
					So(ver.Modified, ShouldBeFalse)
					ver.SetModified()
					So(ver.Modified, ShouldBeTrue)
				})
				Convey("IsModified reflects modified state", func() {
					ver.Modified = false
					So(ver.IsModified(), ShouldBeFalse)
					ver.Modified = true
					So(ver.IsModified(), ShouldBeTrue)
				})
				Convey("ReleaseTable can fail to extract release values", func() {
					_, err := ver.ReleaseTable()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "cannot extract release table")
				})
				Convey("ReleaseTable can fail to get chart data", func() {
					ver.Chart = nil
					_, err := ver.ReleaseTable()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "does not exist in version")
				})
				Convey("ReleaseTable can extract release values", func() {
					ver.Chart = &chart.Chart{
						Values: &chart.Config{
							Values: map[string]*chart.Value{
								"someRelease": &chart.Value{
									Value: "5",
								},
							},
						},
					}
					releaseTable, err := ver.ReleaseTable()
					So(err, ShouldBeNil)
					So(releaseTable, ShouldContainKey, "someRelease")
					So(releaseTable["someRelease"].Value, ShouldEqual, "5")
				})
			})

		})
	})

	Convey("GetVersionsFromList", t, func() {
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)

		// Items will need to be populated with pkg/cluster/driver/util.go:decodeRelease()
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil)

		manifestNames := []string{manifestName}
		versions, err := s.GetVersionsFromList(&manifestNames)
		So(err, ShouldBeNil)
		So(versions, ShouldHaveLength, 1)
		So(versions[0].Name, ShouldEqual, "testManifest")
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
	})

}
