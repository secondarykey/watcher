package main

import (
	"flag"
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var target string
var ignoreS []string
var cmd string
var cmd_argS []string

var version = "0.2.0"
var show_version = flag.Bool("version", false, "show version")

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	var ignore string
	path, _ := os.Getwd()
	flag.StringVar(&target, "target", path, "Search path")
	flag.StringVar(&ignore, "ignore", "tmp;cache;.swp", "Ignore search path(regxp)")
	flag.Parse()

	if *show_version {
		fmt.Printf("version: %s\n", version)
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Arg Error")
		return
	}

	cmds := strings.Split(args[0], " ")

	cmd = cmds[0]
	cmd_argS = cmds[1:]

	ignoreS = strings.Split(ignore, ";")
	run()
}

func run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go monitor(watcher)

	err = watcher.Add(target)
	if err != nil {
		panic(err)
	}
	<-done
}

func monitor(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if !ignore(event.Name) {
					log.Println("modified file:", event.Name)
					command()
				}
			}

		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func ignore(triger string) bool {
	for _, elm := range ignoreS {
		match, _ := regexp.MatchString(elm, triger)
		if match {
			return true
		}
	}
	return false
}

func command() {
	out, _ := exec.Command(cmd, cmd_argS...).CombinedOutput()
	log.Println(string(out))
}
