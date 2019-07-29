package e2e

import (
	"path/filepath"
	"os"
	"os/exec"
	"testing"
	"strconv"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "N" || accUserValue == "n" {
                t.Skip("Skipping Apply Test")
        }

	podNS := "example-go-web-service"
	expectedPodCountForManifest := 1
	expectedPodCountForManifestUpdated := 3
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))
	barrelmanPath, err := filepath.Abs("../barrelman")
	if err != nil {
		t.Log("Absolute path not found for barrelman:", err)
	}
	Convey("Given a manifest", t, func() {
		Convey("When apply is run", func() {
			out, err := exec.Command(barrelmanPath, "apply", "testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")
			Convey("The pod count should be 1", func() {
				So(retryUntilExpectedPodCount(retryCount, podNS, expectedPodCountForManifest), ShouldBeNil)
			})
		})
	})

	Convey("Given an updated manifest", t, func() {
		Convey("When apply is run", func() {
	                out, err := exec.Command(barrelmanPath, "apply", "testdata/manifest_update.yaml").CombinedOutput()
	                So(err, ShouldBeNil)
		        So(string(out), ShouldContainSubstring, "Barrelman")

		        Convey("The pod count should be 3", func() {
	                        So(retryUntilExpectedPodCount(retryCount, podNS, expectedPodCountForManifestUpdated), ShouldBeNil)
	                })
	        })
	})
}
