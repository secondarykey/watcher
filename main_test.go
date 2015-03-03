package main

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
}

func TestGetListDir(t *testing.T) {
}

func TestMonitor(t *testing.T) {
}

func TestNotify(t *testing.T) {
}

func TestIgnore(t *testing.T) {
}

func TestCommand(t *testing.T) {
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
