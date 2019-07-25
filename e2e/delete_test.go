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

	Convey("For the applied  manifest", t, func() {
		Convey("When delete is run", func() {
			os.Chdir("testdata")
	                out, err := exec.Command(barrelmanPath+"/../barrelman", "delete", "manifest_update.yaml").CombinedOutput()
			So(err, ShouldBeNil)
	                So(string(out), ShouldContainSubstring, "deleting release")

			Convey("The pod count should be 0", func() {
				So(WaitForPodsRunningState(podNS, 0), ShouldBeNil)
			})
		})
	})
}
