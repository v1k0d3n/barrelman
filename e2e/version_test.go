package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBarrelmanCommand(t *testing.T) {
	// t.Skip("test")
	podNS := "example-go-web-service"
	//podName := "example-go-web-service"
	podCount := 1
	updatedPodCount := 3
	barrelmanPath := fullPath()
	Convey("Barrelman Binary Exists", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "version").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

	Convey("Barrelman Apply Sample Manifest", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "./testdata/manifest.yaml").CombinedOutput()
		So(err, ShouldBeError)
		fmt.Fprintln(os.Stdout, string(out))
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

	Convey("Check Pod For Applied Manifest By Barrelman", t, func() {
		So(WaitForPodsToBeInRunningState(podNS, podCount), ShouldBeError)
		fmt.Fprintln(os.Stdout, "Pods are in running state")
	})

	Convey("Barrelman Apply Sample Manifest", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "./testdata/manifest_update.yaml").CombinedOutput()
		So(err, ShouldBeError)
		fmt.Fprintln(os.Stdout, string(out))
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

	Convey("Check Pods For Applied Updated Manifest By Barrelman", t, func() {
		So(WaitForPodsToBeInRunningState(podNS, updatedPodCount), ShouldNotPanic)
		fmt.Fprintln(os.Stdout, "Pods are in running state")
	})

}
