package cluster

import (
	"testing"

	yaml "gopkg.in/yaml.v2"

	"google.golang.org/grpc"

	"github.com/charter-se/structured/errors"
	"github.com/stretchr/testify/mock"

	. "github.com/smartystreets/goconvey/convey"
	hapi_chart3 "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
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
	s := NewMockSession()
	Convey("InstallRelease", t, func() {
		Convey("Can fail", func() {
			TestHelm.On("InstallRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(&rls.InstallReleaseResponse{},
				errors.New("Sucessfully failed")).Once()
			_, _, err := s.InstallRelease(&ReleaseMeta{}, []byte{})
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
				},
			}
			TestHelm.On("InstallRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(r, nil).Once()
			_, _, err := s.InstallRelease(&ReleaseMeta{}, []byte{})
			So(err, ShouldBeNil)
		})
	})
}
func TestUpgradeRelease(t *testing.T) {
	s := NewMockSession()
	Convey("UpgradeRelease", t, func() {
		Convey("Can fail", func() {
			TestHelm.On("UpdateRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(&rls.UpdateReleaseResponse{},
				errors.New("Sucessfully failed")).Once()
			_, err := s.UpgradeRelease(&ReleaseMeta{})
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
				},
			}
			TestHelm.On("UpdateRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(r, nil).Once()
			_, err := s.UpgradeRelease(&ReleaseMeta{})
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
			).Return(&rls.UninstallReleaseResponse{},
				grpc.Errorf(grpc.Code(grpc.ErrServerStopped), "Failure sucessful")).Once()
			err := s.DeleteReleases([]*DeleteMeta{
				{
					Name:      "this",
					Namespace: "here",
				},
			})
			So(err, ShouldNotBeNil)
			Print(err)
		})
		Convey("Can succeeed", func() {
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
			).Return(r, nil).Twice()
			err := s.DeleteReleases([]*DeleteMeta{
				{
					Name:      "this1",
					Namespace: "here1",
				},
				{
					Name:      "this2",
					Namespace: "here2",
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
					Name: "something",
					Info: &release.Info{
						Status: &release.Status{
							Code: release.Status_DEPLOYED,
						},
					},
					Chart: &hapi_chart3.Chart{
						Metadata: &hapi_chart3.Metadata{
							Name: "this_chartname",
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
			So(m, ShouldContainKey, "this_chartname")
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
	}

	Convey("DiffRelease", t, func() {
		Convey("Can fail to ReleaseContent", func() {
			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: nil,
			}, errors.New("ReleaseContent should fail")).Once()
			_, _, err := s.DiffRelease(&ReleaseMeta{
				Name:      "something",
				Namespace: "that_namespace",
			})
			So(err, ShouldNotBeNil)
		})
		Convey("Can fail to UpdateRelease (dry run)", func() {
			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: hapiRelease,
			}, nil).Once()

			TestHelm.On("UpdateRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil, errors.New("UpdateRelease should fail")).Once()

			_, _, err := s.DiffRelease(&ReleaseMeta{
				Name:      "something",
				Namespace: "that_namespace",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Failed to get results")
		})

		Convey("Can succeed", func() {

			updateReleaseResp := &rls.UpdateReleaseResponse{
				Release: &hapi_release5.Release{
					Name:     "something",
					Manifest: "\n---\n" + string(newRelease),
				},
			}

			TestHelm.On("ReleaseContent", mock.Anything).Return(&rls.GetReleaseContentResponse{
				Release: hapiRelease,
			}, nil).Once()
			TestHelm.On("UpdateRelease",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(updateReleaseResp, nil).Once()
			changed, bytes, err := s.DiffRelease(&ReleaseMeta{
				Name:      "something",
				Namespace: "that_namespace",
			})
			So(err, ShouldBeNil)
			So(changed, ShouldBeTrue)
			So(string(bytes), ShouldContainSubstring, "Namespace, testRelease, Test (v1) has been removed")
		})
	})
}
