package main

import (
	"github.com/bytedance/gopkg/util/gopool"
	"log"
	"path"
	"sync"
)

type folderCheck folderSync

func makeFolderCheck(diff, srcDir, dstDir string) *folderCheck {
	fs := new(folderCheck)
	fs.diff = diff
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.removed = make(chan string, 64)
	fs.modified = make(chan string, 1024)
	return fs
}

func (fs *folderCheck) Exec() {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for file := range fs.modified {
			src := path.Join(fs.srcDir, file)
			dst := path.Join(fs.dstDir, file)

			wg.Add(1)
			gopool.Go(func() {
				defer wg.Done()

				cmp, err := cmpFile(dst, src)
				if err != nil {
					log.Printf("check file failed, %s, %s", file, err)
					return
				}
				if !cmp {
					log.Printf("The file contents are different: %s, %s", src, dst)
					return
				}
			})
		}
	}()
	go func() {
		defer wg.Done()
		for file := range fs.removed {
			p := path.Join(fs.dstDir, file)
			if exists(p) {
				log.Printf("Failed to delete the file: %s", p)
			}
		}
	}()

	readDiff(fs.diff, fs.modified, fs.removed)
	close(fs.modified)
	close(fs.removed)
	wg.Wait()
}
