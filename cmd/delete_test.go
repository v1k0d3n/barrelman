package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/cluster/mocks"
	"github.com/charter-oss/structured/errors"
)

func TestNewDeleteCmd(t *testing.T) {
	logOpts := []string{}
	Convey("newDeleteCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newDeleteCmd(&barrelman.DeleteCmd{
				Options:    &barrelman.CmdOptions{},
				Config:     &barrelman.Config{},
				LogOptions: &logOpts,
			})
			So(cmd.Name(), ShouldEqual, "delete")
		})
	})
}

func TestDeleteRun(t *testing.T) {
	Convey("Delete", t, func() {
		Convey("Can fail to Init", func() {
			c := &barrelman.DeleteCmd{
				Options: &barrelman.CmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(errors.New("simulated Init failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated Init failure")
			session.AssertExpectations(t)
		})

		Convey("Can fail to resolve charts", func() {
			c := &barrelman.DeleteCmd{
				Options: &barrelman.CmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					ReleaseName: "storage-minio",
					Namespace:   "scratch",
					Chart: &cluster.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, errors.New("simulated ListReleases failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated ListReleases failure")
			session.AssertExpectations(t)
		})

		Convey("Can fail to delete", func() {
			c := &barrelman.DeleteCmd{
				Options: &barrelman.CmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					ReleaseName: "storage-minio",
					Namespace:   "scratch",
					Chart: &cluster.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil).Once()
			session.On("DeleteRelease", mock.Anything).Return(errors.New("simulated delete failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated delete failure")
			session.AssertExpectations(t)
		})

		Convey("Can succeed", func() {
			c := &barrelman.DeleteCmd{
				Options: &barrelman.CmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					ReleaseName: "storage-minio",
					Namespace:   "scratch",
					Chart: &cluster.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil).Once()
			session.On("DeleteRelease", mock.Anything).Return(nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}
