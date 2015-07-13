package main

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setUp()
	os.Exit(m.Run())
	tearDown()
}

func setUp() {
	os.Mkdir("test", 0777)
}

func tearDown() {
	os.Remove("test")
}

func TestRun(t *testing.T) {
}

func TestGetListDir(t *testing.T) {

	gopath := os.Getenv("GOPATH")
	dirs, err := getListDir(gopath + "/src/github.com/secondarykey/watcher")
	if err != nil {
		t.Error("getListDir() Search Error", err)
	}
	if dirs == nil {
		t.Error("Directorys Error")
	}
	if dirs[0] != gopath+"/src/github.com/secondarykey/watcher" {
		t.Error("Directory Search Error", dirs[0])
	}

	dirs, err = getListDir("")
	if err == nil {
		t.Error("getListDir() '' Path Not Error")
	}
	if dirs != nil {
		t.Error("dirs nil Error")
	}

}

func TestMonitor(t *testing.T) {
}

func TestNotify(t *testing.T) {

	//Lock test

}

func TestIgnore(t *testing.T) {
	ignoreS = []string{"aaa", "bbb", "ccc"}
	if ignore("ddd") {
		t.Errorf("ignore() NotFound Error")
	}
	if !ignore("ccc") {
		t.Errorf("ignore() Found Error")
	}
	if ignore("") {
		t.Errorf("ignore() NotFound Error")
	}

}

func ExampleCommand() {

	cmd = "go"
	cmd_argS = []string{"version"}
	command()

	// Output:
	//
	// ********************** command output
	// go version go1.4.1 linux/amd64
	//
	// **********************

}

func ExampleProgress() {

	wait := make(chan bool)
	go progress(wait)

	val := 0
	for {
		select {
		case <-time.After(time.Second * 2):
			val++
			if val == 3 {
				time.Sleep(100 * time.Millisecond)
				wait <- true
				return
			}
		}
	}

	// Output:
	// ###
}
