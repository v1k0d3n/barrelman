package e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanDeleteCommand(t *testing.T) {
	podNS := "example-go-web-service"
	barrelmanPath, _ := os.Getwd()

	Convey("Given a manifest", t, func() {
                out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifest.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		Convey("When delete is run", func() {
	                out, err := exec.Command(barrelmanPath+"/../barrelman", "delete", "testdata/manifest_update.yaml").CombinedOutput()
			So(err, ShouldBeNil)
	                So(string(out), ShouldContainSubstring, "deleting release")

			Convey("The pod count should be 0", func() {
				So(WaitForPodsRunningState(podNS, 0), ShouldBeNil)
			})
		})
	})
}
