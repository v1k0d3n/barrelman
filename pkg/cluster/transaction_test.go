package cluster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTransactions(t *testing.T) {
	// manifestName := "testManifest"
	// s := NewMockSession()

	Convey("Transaction", t, func() {
		Println("Can be created from session")
		// transaction, err := s.NewTransaction(manifestName)
		// So(err, ShouldBeNil)
		// Convey("Can complete", func() {
		// 	err := transaction.Complete()
		// 	So(err, ShouldBeNil)
		// })
	})
}
