package e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	podNS := "example-go-web-service"
	barrelmanPath, _ := os.Getwd()

	Convey("Given a manifest", t, func() {
		Convey("When apply is run", func() {
	                out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeNil)
	                So(string(out), ShouldContainSubstring, "Barrelman")

			Convey("The pod count should be 1", func() {
				So(WaitForPodsRunningState(podNS, 1), ShouldBeNil)
			})
		})
	})

	Convey("Given an updated manifest", t, func() {
                Convey("When apply is run", func() {
                        out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifest_update.yaml").CombinedOutput()
                        So(err, ShouldBeNil)
                        So(string(out), ShouldContainSubstring, "Barrelman")

                        Convey("The pod count should be 3", func() {
                                So(WaitForPodsRunningState(podNS, 3), ShouldBeNil)
                        })
                })
        })
}
