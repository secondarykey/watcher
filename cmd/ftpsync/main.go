package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dutchcoders/goftp"
	"github.com/go-fsnotify/fsnotify"
	"io/ioutil"
	"os"
	_ "os/exec"
	"path/filepath"
	"regexp"
	_ "strings"
	"time"
)

var target string
var ignoreS []string
var ftp *goftp.FTP

var version = "0.0.2"
var show_version = flag.Bool("version", false, "show version")

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

type ftpinfo struct {
	Username   string
	Password   string
	Servername string
	Directory  string
}

func main() {

	path, _ := os.Getwd()
	flag.StringVar(&target, "target", path, "Search path")
	flag.Parse()

	if *show_version {
		fmt.Printf("version: %s\n", version)
		return
	}
	target = "c:/Users/Administrator/Documents/NetBeansProjects/kaigo/kaigo"

	//READ ignore file(.ignore)
	ignoreFile := ".ignore"
	if createIgnore(ignoreFile) {
		fmt.Printf("Loading IgnoreFile: %s\n", ignoreFile)
	} else {
		fmt.Println("Do you may be not read the file?(Y/n):")
		var yn string
		fmt.Scan(&yn)
		if yn != "Y" {
			os.Exit(0)
		}
	}

	ftpFile := ".ftppath"
	info := createFtpfile(ftpFile)
	if info != nil {
		fmt.Printf("Conneting : %s\n", info.Servername)
		if !connectFtp(info) {
			//Remove ?
			fmt.Printf("Connect Error[%s]\n", info.Servername)
			os.Exit(-1)
		}
	} else {
		os.Exit(-1)
	}

	// First upload?
	// Remove FTP File?
	//

	dataMap := make(map[string]string)
	err := getFileMap(dataMap, target)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	//Check Update File(.ftpfile)
	// first upload Y/N
	//Servername

	//files, err := ftp.List("")
	//if err != nil {
	//os.Exit(-1)
	//}
	//fmt.Println(files)

	// func (ftp *FTP) Stor(path string, r io.Reader) (err error)

	// func (ftp *FTP) Retr(path string, retrFn RetrFunc) (s string, err error)

	// ok = ftp.Dele("kaigo/common.php")
	// error(550) = ftp.Dele("kaigo")

	run()
}

func readFileInfo(path string) (map[string]string, error) {
	dataFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()
	dataDecoder := gob.NewDecoder(dataFile)

	newMap := make(map[string]string)
	err = dataDecoder.Decode(&newMap)
	if err != nil {
		return nil, err
	}
	return newMap, nil
}

func writeFileInfo(path string, dataMap map[string]string) error {
	dataFile, err := os.Create(".fileinfo")
	if err != nil {
		return err
	}
	defer dataFile.Close()
	dataEncoder := gob.NewEncoder(dataFile)

	err = dataEncoder.Encode(dataMap)
	if err != nil {
		return err
	}
	return nil
}

func createIgnore(path string) bool {
	fp, err := os.Open(path)
	if err != nil {
		return false
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	ignoreS = make([]string, 30)
	for scanner.Scan() {
		ignoreS = append(ignoreS, scanner.Text())
	}
	return true
}

func createFtpfile(path string) *ftpinfo {
	var info ftpinfo
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("Input FTP Information")
		fmt.Printf("Username:")
		fmt.Scan(&info.Username)
		fmt.Printf("Password:")
		fmt.Scan(&info.Password)

		fmt.Printf("Servername[{ip}:{port}]:")
		fmt.Scan(&info.Servername)

		fmt.Printf("Mapping FTP Directory:")
		fmt.Scan(&info.Directory)

		fmt.Println(info)
		data, err := json.Marshal(info)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		ioutil.WriteFile(path, data, os.ModePerm)
	}

	jsonString, err := ioutil.ReadFile(path)
	err = json.Unmarshal(jsonString, &info)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &info
}

func connectFtp(info *ftpinfo) bool {
	fmt.Println(info)
	var err error
	if ftp, err = goftp.Connect(info.Servername); err != nil {
		fmt.Println("FTP Connect Error")
		fmt.Println(err)
		return false
	}

	if err = ftp.Login(info.Username, info.Password); err != nil {
		fmt.Println("FTP Login Error")
		fmt.Println(err)
		return false
	}

	err = ftp.Cwd(info.Directory)
	if err != nil {
		fmt.Println("Could not Change Directory")
		fmt.Println(err)
		return false
	}

	return true
}

func run() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go monitor(watcher, done)

	fmt.Println("Search start")
	listDirs, err := getListDir(target)
	if err != nil {
		panic(err)
	}
	fmt.Println("Search end")

	for _, elm := range listDirs {
		err = watcher.Add(elm)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Watch:", target)
	<-done
}

func getFileMap(dataMap map[string]string, search string) error {

	fis, err := ioutil.ReadDir(search)
	if err != nil {
		return err
	}

	for _, fi := range fis {

		newPath := filepath.Join(search, fi.Name())
		stamp := fi.ModTime().Format("2006/01/02 15:04:05 MST")

		dataMap[newPath] = stamp
		if !fi.IsDir() {
			continue
		}

		err := getFileMap(dataMap, newPath)
		if err != nil {
			return err
		}
	}
	return nil
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
			fmt.Println("error:", err)
			done <- false
			return
		}
	}
}

func notify(event fsnotify.Event) {

	fmt.Println(event.Name)
	if ignore(event.Name) {
		return
	}

	if event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename ||
		event.Op&fsnotify.Chmod == fsnotify.Chmod {
		fmt.Println(event.Name, event)
		ftpcheck()
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

func ftpcheck() {
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
