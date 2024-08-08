package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/bytedance/gopkg/util/gopool"
)

type folderCheckDir struct {
	ignore []string
	srcDir string
	dstDir string
	files  chan string
	wg     sync.WaitGroup
}

func makeFolderCheckDir(ignoreTxt, srcDir, dstDir string) *folderCheckDir {
	fs := new(folderCheckDir)
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.files = make(chan string, 4096)

	var err error
	fs.ignore, err = readLines(ignoreTxt)
	if err != nil {
		return nil
	}

	return fs
}

func (fs *folderCheckDir) canIgnore(path string) bool {
	for _, reg := range fs.ignore {
		if reg == "" {
			continue
		}
		match, _ := regexp.MatchString(reg, path)
		if match {
			return true
		}
	}
	return false
}

func (fs *folderCheckDir) checkDiff(src string) {
	fs.wg.Add(1)
	gopool.Go(func() {
		defer fs.wg.Done()

		file, err := filepath.Rel(fs.srcDir, src)
		if err != nil {
			log.Printf("get path:%s|%s rel failed, %s", fs.srcDir, src, err)
			return
		}

		if fs.canIgnore(file) {
			return
		}

		dst := path.Join(fs.dstDir, file)
		cmp, err := cmpFile(dst, src)
		if err == nil && cmp {
			return
		}
		copyFile(dst, src)
		log.Printf("fix file: %s", file)
	})
}

func (fs *folderCheckDir) Exec() {
	fs.wg.Add(1)
	go func() {
		defer fs.wg.Done()
		for src := range fs.files {
			fs.checkDiff(src)
		}
	}()

	WalkDir(fs.srcDir, fs.files)
	close(fs.files)
	fs.wg.Wait()
}

func outputCheckDir(output string, filesResult chan string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	for result := range filesResult {
		file.Write([]byte(result))
		file.Write([]byte("\n"))
	}
	return nil
}
