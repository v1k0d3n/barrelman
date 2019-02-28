package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/charter-se/barrelman/pkg/barrelman"
	"github.com/charter-se/barrelman/pkg/cluster"
	"github.com/charter-se/barrelman/pkg/cluster/mocks"
	"github.com/charter-se/structured/errors"
)

func TestNewListCmd(t *testing.T) {
	logOpts := []string{}
	Convey("newListCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newListCmd(&barrelman.ListCmd{
				Options:    &barrelman.CmdOptions{},
				Config:     &barrelman.Config{},
				LogOptions: &logOpts,
			})
			So(cmd.Name(), ShouldEqual, "list")
		})
		Convey("Can fail Run", func() {
			cmd := newListCmd(&barrelman.ListCmd{
				Options:    &barrelman.CmdOptions{},
				Config:     &barrelman.Config{},
				LogOptions: &logOpts,
			})

			err := cmd.RunE(cmd, []string{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "config file does not exist")
		})
	})

}

func TestListRun(t *testing.T) {
	Convey("List", t, func() {
		Convey("Can fail to find config file", func() {
			c := &barrelman.ListCmd{
				Options: &barrelman.CmdOptions{
					ConfigFile: "",
				},
			}
			session := &mocks.Sessioner{}
			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "config file does not exist")
		})
		Convey("Can fail to Init", func() {
			c := &barrelman.ListCmd{
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
		Convey("Can fail to list releases", func() {
			c := &barrelman.ListCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ReleaseName: "storage-minio",
				},
			}, errors.New("simulated Releases failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated Releases failure")
			session.AssertExpectations(t)
		})
		Convey("Can succeed", func() {
			c := &barrelman.ListCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ReleaseName: "storage-minio",
				},
			}, nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}
