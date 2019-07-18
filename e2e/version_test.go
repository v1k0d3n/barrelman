package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBarrelmanCommand(t *testing.T) {
	podNS := "example-go-web-service"
	//podName := "example-go-web-service"
	podCount := 1
	updatedPodCount := 3
	barrelmanPath := fullPath()
	Convey("Barrelman Binary Exists", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "version").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		fmt.Fprintln(os.Stdout, string(out))
	})

	Convey("Barrelman Apply:", t, func() {
		Convey("Test on Sample Manifest", func() {
			out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "./testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeError)
			So(string(out), ShouldContainSubstring, "Barrelman")
			fmt.Fprintln(os.Stdout, string(out))
		})

		Convey("Checking Pod Status", func() {
			So(WaitForPodsToBeInRunningState(podNS, podCount), ShouldBeError)
			fmt.Fprintln(os.Stdout, "Pods are in running state")
		})

		Convey("Test On Sample Updated Manifest With Increased Number Of Replicas", func() {
			out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "./testdata/manifest_update.yaml").CombinedOutput()
			So(err, ShouldBeError)
			So(string(out), ShouldContainSubstring, "Barrelman")
			fmt.Fprintln(os.Stdout, string(out))
		})

		Convey("Check Pods For The Updated Manifest", func() {
			So(WaitForPodsToBeInRunningState(podNS, updatedPodCount), ShouldNotPanic)
			fmt.Fprintln(os.Stdout, "Pods are in running state")
		})
	})
}
