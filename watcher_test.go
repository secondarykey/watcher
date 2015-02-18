package watcher

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testIsSpike(t *testing.T) {
	Convey("Given some integer with a starting value", t, func() {
		x := 1
		Convey("When the integer is incremented", func() {
			x++

			Convey("The value should be greater by one", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}

func testIsUpdateFileInfo(t *testing.T) {
}

func testListFiles(t *testing.T) {
}

func testSearch(t *testing.T) {
}

func testRun(t *testing.T) {
}

func benchmarkListFiles(b *testing.B) {
}
