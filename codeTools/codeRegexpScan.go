package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const (
	EXCLUDE_TEMPLETE = "(vendor)|(.git)|(test)|(generators)"
)

type Scan struct {
	DirPath   string
	Files     []string
	RegLine   *regexp.Regexp
	RegPath   *regexp.Regexp
	WaitGroup sync.WaitGroup
}

type Scaner interface {
	PrintFile()
	Match(field string)
}

func NewScan(dirPath string) Scaner {
	return &Scan{
		Files:   make([]string, 0),
		DirPath: strings.TrimSuffix(dirPath, "/"),
		RegPath: regexp.MustCompile(EXCLUDE_TEMPLETE),
	}
}

func (self *Scan) Match(field string) {
	self.getAllFilesFromDir(self.DirPath)
	self.RegLine = regexp.MustCompile(field)

	for _, file := range self.Files {
		self.WaitGroup.Add(1)
		go self.matchHandle(file)
	}

	self.WaitGroup.Wait()
}

func (self *Scan) PrintFile() {
	for _, file := range self.Files {
		fmt.Println(file)
	}
}

func (self *Scan) getAllFilesFromDir(dirPath string) {
	dirs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("read dir error: %v\n", err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			if !self.RegPath.MatchString(dir.Name()) {
				self.getAllFilesFromDir(dirPath + pathSep + dir.Name())
			}
		} else {
			if !self.RegPath.MatchString(dir.Name()) && strings.HasSuffix(dir.Name(), ".go") {
				self.Files = append(self.Files, dirPath+pathSep+dir.Name())
			}
		}
	}
}

func (self *Scan) matchHandle(filePath string) {
	var pos int64
	fi, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("read file %s error: %v", filePath, err)
		return
	}
	defer func() {
		fi.Close()
		self.WaitGroup.Done()
	}()

	buf := bufio.NewReader(fi)
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		pos++
		if self.RegLine.Match(line) {
			fmt.Printf("%s:%d --> %s\n", filePath, pos, line)
		}
	}
}

var (
	pathSep = string(os.PathSeparator)
	dir     = flag.String("d", "", "project directory")
	reg     = flag.String("r", "", "regexp string")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if *dir == "" || *reg == "" {
		flag.PrintDefaults()
		return
	}

	NewScan(*dir).Match(*reg)
}
