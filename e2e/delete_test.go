package e2e

import (
	"os"
	"os/exec"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanDeleteCommand(t *testing.T) {
	podNS := "example-go-web-service-for-delete"
	barrelmanPath, _ := os.Getwd()
	expectedPodCountForDelete := 0
	retryCount, _ := strconv.Atoi(os.Getenv("RETRYCOUNTACC"))
	BM_TEST_E2E, _ := strconv.ParseBool(os.Getenv("BM_TEST_E2E"))
	if BM_TEST_E2E==true {
		Convey("Given a manifest", t, func() {
			out, err := exec.Command(barrelmanPath+"/../barrelman", "apply", "testdata/manifestForDeleteOps.yaml").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "Barrelman")
			Convey("When delete is run", func() {
				out, err := exec.Command(barrelmanPath+"/../barrelman", "delete", "testdata/manifestForDeleteOps.yaml").CombinedOutput()
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "deleting release")

				Convey("The pod count should be 0", func() {
					So(retryUntilExpectedPodCount(retryCount, podNS, expectedPodCountForDelete), ShouldBeNil)
				})
			})
		})
	} else {
		t.Skip("Skipping Delete Test")
	}
}
