package watcher

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

func TestIsSpike(t *testing.T) {

	Convey("Initialize", t, func() {

		targetFiles = make(map[string]os.FileInfo)
		mode = true
		spike = make(chan string)

		fileName := "work.test"
		os.Create(fileName)
		defer os.Remove(fileName)

		fInfo, _ := os.Stat(fileName)
		targetFiles[fileName] = fInfo

		Convey("mode true", func() {
			So(isSpike(fileName, fInfo), ShouldBeFalse)
		})

		Convey("mode false", func() {
			mode = false
			So(isSpike(fileName, fInfo), ShouldBeFalse)
			Convey("modiefied", func() {

				ctime := time.Now().Local()

				os.Chtimes(fileName, ctime, ctime)
				newInfo, _ := os.Stat(fileName)
				So(isSpike(fileName, newInfo), ShouldBeTrue)
				targetFiles[fileName] = newInfo
				So(isSpike(fileName, newInfo), ShouldBeFalse)
			})

			Convey("New File", func() {
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
