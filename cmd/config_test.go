package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("config", t, func() {
		Convey("Can succeed from file", func() {
			c, err := GetConfigFromFile(getTestDataDir() + "/config")
			So(err, ShouldBeNil)
			So(c, ShouldNotBeNil)
			So(c.Account, ShouldContainKey, "github.com")
		})
		Convey("Can fail from file", func() {
			_, err := GetConfigFromFile(getTestDataDir() + "/unit-test-manifest.yaml")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unit-test-manifest.yaml")
		})
	})
}
