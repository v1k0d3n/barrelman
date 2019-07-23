package e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	podNS := "example-go-web-service"
	podCount := 1
	updatedPodCount := 3
	barrelmanPath, _ := os.Getwd()
	Convey("Testing Using Sample Manifest", t, func() {
		os.Chdir("testdata")
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "manifest.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

	Convey("Testing If Applied Pod Is Running", t, func() {
		So(WaitForPodsRunningState(podNS, podCount), ShouldBeNil)
	})

	Convey("Testing Apply With The Updated Manifest By Increasing Number Of Replicas", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "manifest_update.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

	Convey("Testing If The Applied New Pods Of The Updated Manifest Are In Running State", t, func() {
		So(WaitForPodsRunningState(podNS, updatedPodCount), ShouldBeNil)
	})
}
