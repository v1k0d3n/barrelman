package e2e

import (
	"os"
	"os/exec"
	"testing"
	"strconv"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	podNS := "example-go-web-service"
	barrelmanPath, _ := os.Getwd()
	expectedPodCountForManifest := 1
	expectedPodCountForManifestUpdated := 3
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "N" || accUserValue == "n" {
		t.Skip("Skipping Apply Test")
	}
	Convey("Given a manifest", t, func() {
		Convey("When apply is run", func() {
			out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")
			Convey("The pod count should be 1", func() {
				So(retryUntilExpectedPodCount(retryCount, podNS, expectedPodCountForManifest), ShouldBeNil)
			})
		})
	})

	Convey("Given an updated manifest", t, func() {
		Convey("When apply is run", func() {
	                out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifest_update.yaml").CombinedOutput()
	                So(err, ShouldBeNil)
		        So(string(out), ShouldContainSubstring, "Barrelman")

		        Convey("The pod count should be 3", func() {
	                        So(retryUntilExpectedPodCount(retryCount, podNS, expectedPodCountForManifestUpdated), ShouldBeNil)
	                })
	        })
	})
}
