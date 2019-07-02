package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaults(t *testing.T) {
	Convey("newDefaults", t, func() {
		d := Default()
		Convey("Can has ConfigFile", func() {
			So(ConfigFile, ShouldNotBeEmpty)
		})
		Convey("Can has ManifestFile", func() {
			So(ManifestFile, ShouldNotBeEmpty)
		})
		Convey("Can has KubeConfigFile", func() {
			So(KubeConfigFile, ShouldNotBeEmpty)
		})
		//KubeContext is empty by default unless KUBE_CONTEXT is set
		Convey("Can has DataDir", func() {
			So(DataDir, ShouldNotBeEmpty)
		})
		Convey("Can has InstallRetry", func() {
			So(InstallRetry, ShouldNotBeEmpty)
		})
		Convey("Can has Force", func() {
			So(Force, ShouldNotBeNil)
		})
	})
}
