package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaults(t *testing.T) {
	Convey("newDefaults", t, func() {
		d := Default()
		Convey("Can has ConfigFile", func() {
			So(d.ConfigFile, ShouldNotBeEmpty)
		})
		Convey("Can has ManifestFile", func() {
			So(d.ManifestFile, ShouldNotBeEmpty)
		})
		Convey("Can has KubeConfigFile", func() {
			So(d.KubeConfigFile, ShouldNotBeEmpty)
		})
		//KubeContext is empty by default unless KUBE_CONTEXT is set
		Convey("Can has DataDir", func() {
			So(d.DataDir, ShouldNotBeEmpty)
		})
		Convey("Can has InstallRetry", func() {
			So(d.InstallRetry, ShouldNotBeEmpty)
		})
		Convey("Can has Force", func() {
			So(d.Force, ShouldNotBeNil)
		})
	})
}
