package main

import (
	"flag"
	"fmt"
	"github.com/secondarykey/watcher"
	"os"
)

var version = "0.1.0"
var show_version = flag.Bool("version", false, "show version")
var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	path, _ := os.Getwd()

	var target string
	var duration int
	var ignore string

	patrol := flag.Bool("patrol", false, "Show Patrol TimeStamp")
	flag.StringVar(&target, "target", path, "Search path")
	flag.IntVar(&duration, "duration", 10, "Search duration")
	flag.StringVar(&ignore, "ignore", "cache", "Search path")
	flag.Parse()

	args := flag.Args()
	if *show_version {
		fmt.Printf("version: %s\n", version)
		return
	}

	if len(args) != 1 {
		fmt.Println("Arg Error")
		return
	}

	watcher.Run(target, ignore, args[0], duration, *patrol)
}
