package cmd

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/cluster/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestNewApplyCmd(t *testing.T) {
	Convey("newApplyCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newApplyCmd(&applyCmd{
				Options: &cmdOptions{},
				Config:  &Config{},
			})
			So(cmd.Name(), ShouldEqual, "apply")
		})
	})
}

func TestApplyRun(t *testing.T) {
	Convey("Run", t, func() {
		//cwd, err := osext.ExecutableFolder()
		//So(err, ShouldBeNil)
		Convey("Can fail to find config file", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile: "",
				},
			}
			session := &mocks.Sessioner{}
			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})

		Convey("Can succeed with one install failure (retry)", func() {
			c := &applyCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{}, nil).Once()

			//This run of InstallRelease will be a dry run to check for errors
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			//This run will be the real deal, but we error first round
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install failed", "some_release", errors.New("Error injection")).Once()

			//This run will be the real deal
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
		})

		Convey("Should fail after retry count exceeded", func() {
			c := &applyCmd{
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
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{}, nil).Once()

			//This run of InstallRelease will be a dry run to check for errors
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			//This run will be the real deal, but we error first round
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install failed", "some_release", errors.New("Error injection")).Times(3)

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Error injection")
		})
	})
}

//getTestDataDir returns a string representing the location of the testdata directory as derived from THIS source file
//our tests are run in temporary directories, so finding the testdata can be a little troublesome
func getTestDataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return fmt.Sprintf("%v/../testdata", path.Dir(filename))
}
