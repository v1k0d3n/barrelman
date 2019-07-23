package e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	podNS := "example-go-web-service"
	//podName := "example-go-web-service"
	podCount := 1
	updatedPodCount := 3
	barrelmanPath, _ := os.Getwd()
	Convey("Testing using Sample Manifest", t, func() {
		os.Chdir("testdata")
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "manifest.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		t.Log(string(out))
	})

	Convey("Checking Pod Status", t, func() {
		So(WaitForPodsToBeInRunningState(podNS, podCount), ShouldBeNil)
		t.Log("Pods are in running state")
	})

	Convey("Testing With The Increased Number Of Replicas", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "manifest_update.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		t.Log(string(out))
	})

	Convey("Checking Number Of Pods As Per The Updated Manifest", t, func() {
		So(WaitForPodsToBeInRunningState(podNS, updatedPodCount), ShouldBeNil)
		t.Log("Pods are in running state")
	})
}
