package e2e

import (
	"os"
	"os/exec"
	"testing"
	"strconv"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "" || accUserValue == "n" {
                t.Log("To run Acceptance tests, run 'BM_TEST_E2E=y BM_BIN=[PathOfBarrelman] RETRYCOUNTACC=20 go test ./e2e -v'")
                t.Skip("Skipping Apply Test")
        }
        bmBin := os.Getenv("BM_BIN")
        if bmBin == "" {
                t.Fatal("Barrelman reference path is not set in BM_BIN parameter")
        }
	podNS := "example-go-web-service"
	podName := "example-go-web-service"
	expectedPodCountForManifest := 1
	expectedPodCountForManifestUpdated := 3
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))

	Convey("Given a manifest", t, func() {
		Convey("When apply is run", func() {
			out, err := exec.Command(bmBin, "apply", "testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")
			Convey("The pod count should be 1", func() {
				So(retryUntilExpectedPodCount(retryCount, podNS, podName, expectedPodCountForManifest), ShouldBeNil)

			})
		})
	})

	Convey("Given an updated manifest", t, func() {
		Convey("When apply is run", func() {
	                out, err := exec.Command(bmBin, "apply", "testdata/manifest_update.yaml").CombinedOutput()
	                So(err, ShouldBeNil)
		        So(string(out), ShouldContainSubstring, "Barrelman")

		        Convey("The pod count should be 3", func() {
	                        So(retryUntilExpectedPodCount(retryCount, podNS, podName, expectedPodCountForManifestUpdated), ShouldBeNil)
	                })
	        })
	})
}
