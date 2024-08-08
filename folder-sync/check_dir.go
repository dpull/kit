package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/bytedance/gopkg/util/gopool"
)

type folderCheckDir struct {
	srcDir string
	dstDir string
	files  chan string
	result chan string
	wg     sync.WaitGroup
}

func makeFolderCheckDir(srcDir, dstDir string) *folderCheckDir {
	fs := new(folderCheckDir)
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.files = make(chan string, 4096)
	fs.result = make(chan string, 1024)
	return fs
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

		dst := path.Join(fs.dstDir, file)
		cmp, err := cmpFile(dst, src)
		if err != nil {
			fs.result <- fmt.Sprintf("A\t%s", file)
			return
		}
		if !cmp {
			fs.result <- fmt.Sprintf("M\t%s", file)
			return
		}
	})
}

func (fs *folderCheckDir) Exec() {
	fs.wg.Add(2)
	go func() {
		defer fs.wg.Done()
		for src := range fs.files {
			fs.checkDiff(src)
		}
		close(fs.result)
	}()

	go func() {
		defer fs.wg.Done()
		outputCheckDir("check-dir.csv", fs.result)
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
