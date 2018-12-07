package cmd

import (
	"testing"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/cluster/mocks"
	"github.com/charter-se/structured/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewListCmd(t *testing.T) {
	Convey("newListCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newListCmd(&listCmd{
				Options: &cmdOptions{},
				Config:  &Config{},
			})
			So(cmd.Name(), ShouldEqual, "list")
		})
		Convey("Can fail Run", func() {
			cmd := newListCmd(&listCmd{
				Options: &cmdOptions{},
				Config:  &Config{},
			})

			err := cmd.RunE(cmd, []string{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})
	})

}

func TestListRun(t *testing.T) {
	Convey("List", t, func() {
		Convey("Can fail to find config file", func() {
			c := &listCmd{
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
			c := &listCmd{
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
		Convey("Can fail to list releases", func() {
			c := &listCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Name: "storage-minio",
				},
			}, errors.New("simulated Releases failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated Releases failure")
			session.AssertExpectations(t)
		})
		Convey("Can succeed", func() {
			c := &listCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Name: "storage-minio",
				},
			}, nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}
