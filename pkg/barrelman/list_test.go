package barrelman

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/charter-se/barrelman/pkg/cluster"
	"github.com/charter-se/barrelman/pkg/cluster/mocks"
	"github.com/charter-se/barrelman/pkg/manifest/chartsync"
	"github.com/charter-se/structured/errors"
)

func TestListCmd(t *testing.T) {
	newListCmd := func() *ListCmd {
		return &ListCmd{
			Options: &CmdOptions{
				Force:          &[]string{},
				DataDir:        "testdata/datadir",
				ConfigFile:     "testdata/config",
				KubeConfigFile: "testdata/kubeconfig",
				KubeContext:    "default",
				ManifestFile:   "testdata/manifest.yaml",
			},
			Config: &Config{
				Account: make(chartsync.AccountTable),
			},
		}
	}

	Convey("List Command", t, func() {
		listCmd := newListCmd()
		session := &mocks.Sessioner{}

		Convey("Should error on session.Init()", func() {
			session.On("Init").Return(errors.New("simulated Init error"))
			err := listCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should error on session.Releases()", func() {
			releases := make(map[string]*cluster.ReleaseMeta)
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(listCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(listCmd.Options.KubeContext)
			session.On("Releases").Return(releases, errors.New("simulated"))
			err := listCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should succeed", func() {
			releases := make(map[string]*cluster.ReleaseMeta)
			releases["someRelease"] = &cluster.ReleaseMeta{
				ReleaseName: "simulated-release",
			}
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(listCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(listCmd.Options.KubeContext)
			session.On("Releases").Return(releases, nil)
			err := listCmd.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}