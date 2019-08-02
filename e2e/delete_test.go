package e2e

import (
	"os"
	"os/exec"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanDeleteCommand(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "" {
		t.Log("To run Acceptance tests, run 'make testacc")
		t.Skip("Skipping Delete Test")
	}
	bmBin := os.Getenv("BM_BIN")
	if bmBin == "" {
		t.Fatal("Barrelman binary environment variable BM_BIN not set")
	}
	podNS := "example-go-web-service-for-delete"
	expectedPodCountForDelete := 0
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))
	interval, _ := strconv.Atoi(os.Getenv("INTERVALTIME"))
	Convey("Given a manifest", t, func() {
		out, err := exec.Command(bmBin, "apply", "testdata/manifestForDeleteOps.yaml").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		Convey("When delete is run", func() {
			out, err := exec.Command(bmBin, "delete", "testdata/manifestForDeleteOps.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "deleting release")

			Convey("The pod count should be 0", func() {
				f := func() error {
					return checkPodCount(podNS, podName, expectedPodCountForDelete)
				}
				So(retry(f, retryCount, interval, []string{"WrongNumberOfPods"}), ShouldBeNil)
			})
		})
	})
}
