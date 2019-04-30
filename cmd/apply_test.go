package cmd

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster/mocks"
)

func TestNewApplyCmd(t *testing.T) {
	Convey("newApplyCmd", t, func() {
		logOpts := []string{}
		Convey("Can succeed", func() {
			cmd := newApplyCmd(&barrelman.ApplyCmd{
				Options:    &barrelman.CmdOptions{},
				Config:     &barrelman.Config{},
				LogOptions: &logOpts,
			})
			So(cmd.Name(), ShouldEqual, "apply")
		})
	})
}

func TestApplyRun(t *testing.T) {
	Convey("Run", t, func() {
		Convey("Can handle Init failure", func() {
			c := &barrelman.ApplyCmd{
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
	})
}

//getTestDataDir returns a string representing the location of the testdata directory as derived from THIS source file
//our tests are run in temporary directories, so finding the testdata can be a little troublesome
func getTestDataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return fmt.Sprintf("%v/../testdata", path.Dir(filename))
}
