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
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var target string
var ftpDir string
var ignoreS []string
var ftp *goftp.FTP
var queue []fsnotify.Event
var queueLock bool

var version = "0.0.2"
var show_version = flag.Bool("version", false, "show version")

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

type queueData struct {
	event fsnotify.Event
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
	target = "/home/secondarykey/work"

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
	_, err := os.Stat(ftpFile)
	var info *ftpinfo
	if err != nil {
		info = createFtpfile(ftpFile)
	} else {
		info = readFtpfile(ftpFile)
	}

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

	ftpDir = info.Directory

	dataMap := make(map[string]string)
	err = getFileMap(dataMap, target)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	files, err := ftp.List(info.Directory)
	if err != nil {
		os.Exit(-1)
	}
	fmt.Println(files)

	// func (ftp *FTP) Stor(path string, r io.Reader) (err error)
	// func (ftp *FTP) Retr(path string, retrFn RetrFunc) (s string, err error)

	// ok = ftp.Dele("kaigo/common.php")
	// error(550) = ftp.Dele("kaigo")
	go queueProcess()

	run()
}

func deleteFTP(target string) error {
	path := ftpDir + target
	fmt.Println("FTP:" + path)
	return ftp.Dele(ftpDir + target)
}

func uploadFTP(target string, full string) error {

	file, err := os.Open(full)
	if err != nil {
		return err
	}
	path := ftpDir + target
	fmt.Println("FTP:" + path)
	return ftp.Stor(path, file)
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
	ignoreS = make([]string, 0, 30)
	for scanner.Scan() {
		ignoreS = append(ignoreS, scanner.Text())
	}
	return true
}

func createFtpfile(path string) *ftpinfo {
	var info ftpinfo
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
	writeFtpfile(path, data)
	return &info
}

func readFtpfile(path string) *ftpinfo {
	var info ftpinfo
	jsonString, err := ioutil.ReadFile(path)
	err = json.Unmarshal(jsonString, &info)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &info
}

func writeFtpfile(path string, data []byte) error {
	ioutil.WriteFile(path, data, os.ModePerm)
	return nil
}

func connectFtp(info *ftpinfo) bool {
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

	_, err = ftp.Pasv()
	if err != nil {
		fmt.Println("Change PASV mode")
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

	if ignore(event.Name) {
		return
	}

	queue = append(queue, event)

	return
}

func newFileName(fullPath string, target string) string {
	newEvt := strings.Replace(fullPath, "\\", "/", -1)
	newStr := strings.Replace(newEvt, target, "", 1)
	return newStr
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

func queueProcess() {

	queue = make([]fsnotify.Event, 0, 100)
	queueLock = false

	for {
		select {
		case <-time.After(time.Second * 3):
			if !queueLock {
				queueLock = true
				dst := renewQueue()
				processFTP(dst)
				fmt.Println("#", len(queue))
				queueLock = false
			}
		}
	}
}

func renewQueue() []fsnotify.Event {
	pointer := queue
	queue = make([]fsnotify.Event, 0, 100)
	return pointer
}

func processFTP(events []fsnotify.Event) {

	for _, event := range events {
		fmt.Println(event)

		if (event.Op&fsnotify.Write == fsnotify.Write) ||
			(event.Op&fsnotify.Create == fsnotify.Create) {

			info, _ := os.Stat(event.Name)
			name := newFileName(event.Name, target)
			if info == nil {
				fmt.Println("info == nil")
				return
			}

			if info.IsDir() {
			} else {
				err := ftp.Type("A")
				fmt.Println(err)

				err = uploadFTP(name, event.Name)
				if err != nil {
					fmt.Println("upload error:", err)
				}
			}

		} else if event.Op&fsnotify.Remove == fsnotify.Remove {

			name := newFileName(event.Name, target)
			err := deleteFTP(name)
			if err != nil {
				fmt.Println("delete error:", err)
			}

			//event.Op&fsnotify.Rename == fsnotify.Rename {
			//event.Op&fsnotify.Chmod == fsnotify.Chmod {
		} else {
			fmt.Println(event)
		}
	}
}
