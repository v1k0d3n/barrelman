package cluster

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
	hapi_chart3 "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/charter-oss/structured/errors"
)

func TestListReleases(t *testing.T) {
	s := NewMockSession()
	Convey("ListReleases", t, func() {
		Convey("Can fail", func() {
			r := []*hapi_release5.Release{
				{
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Config: &hapi_chart3.Config{
						Raw:    "this: that\n",
						Values: make(map[string]*hapi_chart3.Value),
					},
				},
			}

			TestHelm.On("ListReleases", mock.Anything).Return(&rls.ListReleasesResponse{
				Count:    int64(len(r)),
				Releases: r,
			}, errors.New("LsitReleases should fail")).Once()
			_, err := s.ListReleases()
			So(err, ShouldNotBeNil)
		})
		Convey("Can succeed", func() {
			r := []*hapi_release5.Release{
				{
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Config: &hapi_chart3.Config{
						Raw:    "this: that\n",
						Values: make(map[string]*hapi_chart3.Value),
					},
				},
			}

			TestHelm.On("ListReleases", mock.Anything).Return(&rls.ListReleasesResponse{
				Count:    int64(len(r)),
				Releases: r,
			}, nil).Once()
			releases, err := s.ListReleases()
			So(err, ShouldBeNil)
			So(releases, ShouldHaveLength, 1)
			So(releases[0].Status, ShouldEqual, Status_DEPLOYED)
			Convey("The value should be greater by one", func() {

			})
		})
	})
}
func TestInstallRelease(t *testing.T) {
	manifestName := "testGroup"
	s := NewMockSession()
	Convey("InstallRelease", t, func() {
		Convey("Can fail", func() {
			TestHelm.On("InstallReleaseFromChart",
				mock.Anything,
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(&rls.InstallReleaseResponse{},
				errors.New("Sucessfully failed")).Once()
			_, err := s.InstallRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
				Chart: &hapi_chart3.Chart{
					Metadata: &hapi_chart3.Metadata{
						Name: "something",
					},
				},
			}, manifestName)
			So(err, ShouldNotBeNil)
			Print(err)
		})
		Convey("Can succeeed", func() {
			r := &rls.InstallReleaseResponse{
				Release: &hapi_release5.Release{
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Config: &hapi_chart3.Config{
						Raw:    "this: that\n",
						Values: make(map[string]*hapi_chart3.Value),
					},
				},
			}
			TestHelm.On("InstallReleaseFromChart",
				mock.Anything,
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(r, nil).Once()
			_, err := s.InstallRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
				Chart: &hapi_chart3.Chart{
					Metadata: &hapi_chart3.Metadata{
						Name: "something",
					},
				},
			}, manifestName)
			So(err, ShouldBeNil)
		})
	})
}
func TestUpgradeRelease(t *testing.T) {
	manifestName := "testGroup"
	s := NewMockSession()
	Convey("UpgradeRelease", t, func() {
		Convey("Can fail", func() {
			TestHelm.On("UpdateReleaseFromChart",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(&rls.UpdateReleaseResponse{},
				errors.New("Sucessfully failed")).Once()
			_, err := s.UpgradeRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
				Chart: &hapi_chart3.Chart{
					Metadata: &hapi_chart3.Metadata{
						Name: "something",
					},
				},
			}, manifestName)
			So(err, ShouldNotBeNil)
			Print(err)
		})
		Convey("Can succeeed", func() {
			r := &rls.UpdateReleaseResponse{
				Release: &hapi_release5.Release{
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Config: &hapi_chart3.Config{
						Raw:    "this: that\n",
						Values: make(map[string]*hapi_chart3.Value),
					},
				},
			}
			TestHelm.On("UpdateReleaseFromChart",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(r, nil).Once()
			_, err := s.UpgradeRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
				Chart: &hapi_chart3.Chart{
					Metadata: &hapi_chart3.Metadata{
						Name: "something",
					},
				},
			}, manifestName)
			So(err, ShouldBeNil)
		})
	})
}
func TestDeleteRelease(t *testing.T) {
	s := NewMockSession()
	Convey("DeleteReleases", t, func() {
		Convey("Can fail", func() {
			TestHelm.On("DeleteRelease",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
			).Return(&rls.UninstallReleaseResponse{},
				grpc.Errorf(grpc.Code(grpc.ErrServerStopped), "Failure sucessful")).Once()
			err := s.DeleteReleases([]*DeleteMeta{
				{
					ReleaseName: "this",
					Namespace:   "here",
				},
			})
			So(err, ShouldNotBeNil)
			Print(err)
		})
		SkipConvey("Can succeeed", func() {
			r := &rls.UninstallReleaseResponse{
				Release: &hapi_release5.Release{
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
				},
			}
			TestHelm.On("DeleteRelease",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
			).Return(r, nil).Twice()
			err := s.DeleteReleases([]*DeleteMeta{
				{
					ReleaseName: "this1",
					Namespace:   "here1",
				},
				{
					ReleaseName: "this2",
					Namespace:   "here2",
				},
			})
			So(err, ShouldBeNil)
		})
	})
}
func TestReleases(t *testing.T) {
	s := NewMockSession()
	Convey("Releases", t, func() {
		Convey("Can fail", func() {
			r := []*hapi_release5.Release{}
			TestHelm.On("ListReleases", mock.Anything).Return(&rls.ListReleasesResponse{
				Count:    int64(len(r)),
				Releases: r,
			}, errors.New("successfuly failed")).Once()
			_, err := s.Releases()
			So(err, ShouldNotBeNil)
		})
		Convey("Can succeed", func() {
			r := []*hapi_release5.Release{
				{
					Name: "this_chartname",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Chart: &hapi_chart3.Chart{
						Metadata: &hapi_chart3.Metadata{
							Name: "something",
						},
					},
				},
			}
			TestHelm.On("ListReleases", mock.Anything).Return(&rls.ListReleasesResponse{
				Count:    int64(len(r)),
				Releases: r,
			}, nil).Once()
			m, err := s.Releases()
			So(err, ShouldBeNil)
			Print(m)
			So(m, ShouldContainKey, "this_chartname")
		})
	})
}

func TestGetRelease(t *testing.T) {
	releaseName := "thisRelease"
	revision := int32(77)
	s := NewMockSession()
	Convey("Releases", t, func() {
		Convey("Can succeed", func() {
			r := &hapi_release5.Release{
				Name: releaseName,
				Info: &release.Info{
					Status: &release.Status{
						Code: release.Status_DEPLOYED,
					},
				},
				Chart: &hapi_chart3.Chart{
					Metadata: &hapi_chart3.Metadata{
						Name:    releaseName,
						Version: fmt.Sprintf("%d", revision),
					},
				},
			}
			TestHelm.On("ReleaseContent", releaseName, mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: r,
			}, nil).Once()
			release, err := s.GetRelease(releaseName, revision)
			So(err, ShouldBeNil)
			So(release.ReleaseName, ShouldEqual, releaseName)
		})

		Convey("Can fail", func() {
			TestHelm.On("ReleaseContent", releaseName, mock.Anything).Return(nil, errors.New("Sim error")).Once()
			_, err := s.GetRelease(releaseName, revision)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Sim error")
		})

	})
}

func TestDiffRelease(t *testing.T) {
	s := NewMockSession()
	//origRelease serves as the release already deployed on k8s
	origRelease, err := yaml.Marshal(metadata{
		ApiVersion: "v1",
		Kind:       "Test",
		Metadata: struct {
			Namespace string
			Name      string
		}{
			Name:      "testRelease",
			Namespace: "testNamespace",
		},
	})
	if err != nil {
		panic("Could not marshall document")
	}
	//newRelease serves as the response to a dry-run with changes
	newRelease, err := yaml.Marshal(metadata{
		ApiVersion: "v1",
		Kind:       "Test",
		Metadata: struct {
			Namespace string
			Name      string
		}{
			Name:      "testRelease",
			Namespace: "testNamespace2",
		},
	})
	if err != nil {
		panic("Could not marshall document")
	}

	hapiRelease := &hapi_release5.Release{
		Name:     "something",
		Manifest: "\n---\n" + string(origRelease),
		Info: &release.Info{
			Status: &release.Status{
				Code: release.Status_DEPLOYED,
			},
		},
		Config: &hapi_chart3.Config{
			Raw:    "this: that\n",
			Values: make(map[string]*hapi_chart3.Value),
		},
	}

	Convey("DiffRelease", t, func() {
		Convey("Can fail to ReleaseContent", func() {
			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: nil,
			}, errors.New("ReleaseContent should fail")).Once()
			_, _, err := s.DiffRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
			})
			So(err, ShouldNotBeNil)
		})
		Convey("Can fail to UpdateRelease (dry run)", func() {
			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: hapiRelease,
			}, nil).Once()

			TestHelm.On("UpdateReleaseFromChart",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil, errors.New("UpdateRelease should fail")).Once()

			_, _, err := s.DiffRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Failed to get results")
		})

		Convey("Can succeed", func() {

			updateReleaseResp := &rls.UpdateReleaseResponse{
				Release: &hapi_release5.Release{
					Name:     "something",
					Manifest: "\n---\n" + string(newRelease),
					Config: &hapi_chart3.Config{
						Raw:    "this: that\n",
						Values: make(map[string]*hapi_chart3.Value),
					},
				},
			}

			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: hapiRelease,
			}, nil).Once()
			TestHelm.On("UpdateReleaseFromChart",
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(updateReleaseResp, nil).Once()
			changed, bytes, err := s.DiffRelease(&ReleaseMeta{
				ReleaseName: "something",
				Namespace:   "that_namespace",
			})
			So(err, ShouldBeNil)
			So(changed, ShouldBeTrue)
			So(string(bytes), ShouldContainSubstring, "Namespace, testRelease, Test (v1) has been removed")
		})
	})
}
