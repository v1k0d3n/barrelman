package e2e

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestAccBarrelmanApplyCommand(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "" {
		t.Log("To run Acceptance tests, run 'make testacc")
		t.Skip("Skipping Apply Test")
	}
	bmBin := os.Getenv("BM_BIN")
	if bmBin == "" {
		t.Fatal("Barrelman binary environment variable BM_BIN not set")
	}
	podNS := "example-go-web-service"
	podName := "example-go-web-service"
	expectedPodCountForManifest := 1
	expectedPodCountForManifestUpdated := 3
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))
	interval, _ := strconv.Atoi(os.Getenv("INTERVALTIME"))
	Convey("Given a manifest", t, func() {
		Convey("When apply is run", func() {
			out, err := exec.Command(bmBin, "apply", "testdata/manifest.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")
			Convey("The pod count should be 1", func() {
				f := func() error {
					return checkPodCount(podNS, podName, expectedPodCountForManifest)
				}

				So(retry(f, retryCount, interval, []string{"WrongNumberOfPods"}), ShouldBeNil)
			})
		})
	})

	Convey("Given an updated manifest", t, func() {
		Convey("When apply is run", func() {
			out, err := exec.Command(bmBin, "apply", "testdata/manifest_update.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")

			Convey("The pod count should be 3", func() {
				f := func() error {
					return checkPodCount(podNS, podName, expectedPodCountForManifestUpdated)
				}

				So(retry(f, retryCount, interval, []string{"WrongNumberOfPods"}), ShouldBeNil)
			})
		})
	})
}
