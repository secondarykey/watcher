package main

import (
	"flag"
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var target string
var ignoreS []string
var cmd string
var cmd_argS []string
var lock = false

var version = "0.2.1"
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

	go monitor(watcher, done)

	log.Println("Search start")
	listDirs, err := getListDir(target)
	if err != nil {
		panic(err)
	}
	log.Println("Search end")

	for _, elm := range listDirs {
		err = watcher.Add(elm)
		if err != nil {
			panic(err)
		}
	}

	log.Println("Watch:", target)
	<-done
}

func getListDir(search string) ([]string, error) {

	list := make([]string, 0)
	list = append(list, search)

	fis, err := ioutil.ReadDir(search)
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		newSearchPath := filepath.Join(search, fi.Name())
		newList, err := getListDir(newSearchPath)
		if err != nil {
			return nil, err
		}

		list = append(list, newList...)
	}

	return list, nil
}

func monitor(watcher *fsnotify.Watcher, done chan bool) {
	for {
		select {
		case event := <-watcher.Events:
			go notify(event)
		case err := <-watcher.Errors:
			log.Println("error:", err)
			done <- false
			return
		}
	}
}

func notify(event fsnotify.Event) {

	if lock {
		return
	}

	lock = true
	defer func() {
		lock = false
	}()

	if ignore(event.Name) {
		return
	}

	if event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename ||
		event.Op&fsnotify.Chmod == fsnotify.Chmod {
		log.Println(event.Name, event)
		log.Println("********************** running command (", cmd, cmd_argS, ")")
		command()
		log.Println("********************** ending  command")
	}
	return
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
	wait := make(chan bool)
	go progress(wait)
	out, _ := exec.Command(cmd, cmd_argS...).CombinedOutput()
	wait <- true
	fmt.Println("")
	fmt.Println("********************** command output")
	fmt.Println(string(out))
	fmt.Println("**********************")
}

func progress(wait chan bool) {
	for {
		select {
		case <-wait:
			return
		case <-time.After(time.Second * 2):
			fmt.Printf("#")
		}
	}
}
