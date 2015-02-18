package watcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var targetFiles map[string]os.FileInfo
var mode = true
var spike chan string

func Run(target, ignore, cmd string, duration int, patrol bool) {

	fmt.Println("Watcher Start")
	targetFiles = make(map[string]os.FileInfo)
	spike = make(chan string)

	err := search(target)
	if err != nil {
		panic(err)
	}

	cmds := strings.Split(cmd, " ")
	for {
		select {
		case timer := <-time.After(time.Duration(duration) * time.Second):
			if patrol {
				fmt.Println(timer)
			}
			mode = false
			go search(target)
		case triger := <-spike:
			match, _ := regexp.MatchString(ignore, triger)
			if !match {
				fmt.Println(triger)
				out, _ := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
				fmt.Println(string(out))
			}
		}
	}
}

func search(rootPath string) error {
	err := updateFileInfo(rootPath, rootPath)
	return err
}

func listFiles(rootPath, searchPath string) error {

	fis, err := ioutil.ReadDir(searchPath)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		fullPath := filepath.Join(searchPath, fi.Name())
		err := updateFileInfo(rootPath, fullPath)
		if err != nil {
			return nil
		}
	}
	return nil
}

func updateFileInfo(rootPath, path string) error {

	fInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if isSpike(path, fInfo) {
		mode = true
		spike <- path
	}

	// fileinfo update
	targetFiles[path] = fInfo
	if fInfo.IsDir() {
		go listFiles(rootPath, path)
	}
	return nil
}

func isSpike(fullPath string, fInfo os.FileInfo) bool {

	//update mode
	if mode {
		return false
	}

	src := targetFiles[fullPath]
	if src != nil {
		if fInfo.ModTime() != src.ModTime() {
			return true
		}
	} else {
		return true
	}
	return false
}
