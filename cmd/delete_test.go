package cmd

import (
	"testing"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/cluster/mocks"
	"github.com/charter-se/structured/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func TestNewDeleteCmd(t *testing.T) {
	Convey("newDeleteCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newDeleteCmd(&deleteCmd{
				Options: &cmdOptions{},
				Config:  &Config{},
			})
			So(cmd.Name(), ShouldEqual, "delete")
		})
	})
}

func TestDeleteRun(t *testing.T) {
	Convey("Delete", t, func() {
		//cwd, err := osext.ExecutableFolder()
		//So(err, ShouldBeNil)
		Convey("Can fail to find config file", func() {
			c := &deleteCmd{
				Options: &cmdOptions{
					ConfigFile: "",
				},
			}
			session := &mocks.Sessioner{}
			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})

		Convey("Can fail to Init", func() {
			c := &deleteCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
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
			c := &deleteCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Once()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					Name:      "storage-minio",
					Namespace: "scratch",
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
			c := &deleteCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Once()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					Name:      "storage-minio",
					Namespace: "scratch",
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
			c := &deleteCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Once()
			session.On("GetKubeContext").Return("").Once()
			session.On("ListReleases").Return([]*cluster.Release{
				&cluster.Release{
					Name:      "storage-minio",
					Namespace: "scratch",
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
