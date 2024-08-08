package main

import (
	"log"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/bytedance/gopkg/util/gopool"
)

type folderCheckDir folderSync

func makeFolderCheckDir(srcDir, dstDir string) *folderCheckDir {
	fs := new(folderCheckDir)
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.modified = make(chan string, 1024)
	return fs
}

func (fs *folderCheckDir) Exec() {
	var wg sync.WaitGroup
	var diff atomic.Int32
	var exist atomic.Int32

	wg.Add(2)
	go func() {
		defer wg.Done()

		for src := range fs.modified {
			file, err := filepath.Rel(fs.srcDir, src)
			if err != nil {
				log.Printf("get path:%s|%s rel failed, %s", fs.srcDir, src, err)
				continue
			}

			dst := path.Join(fs.dstDir, file)
			wg.Add(1)
			gopool.Go(func() {
				defer wg.Done()

				cmp, err := cmpFile(dst, src)
				if err != nil {
					exist.Add(1)
					log.Printf("check file failed, %s, %s", file, err)
					return
				}
				if !cmp {
					diff.Add(1)
					log.Printf("The file contents are different: %s, %s", src, dst)
					return
				}
			})
		}
	}()

	WalkDir(fs.srcDir, fs.modified)
	close(fs.modified)
	wg.Wait()

	if diff.Load() != 0 {
		log.Printf("The file contents are different count:%d", diff.Load())
	}
	if exist.Load() != 0 {
		log.Printf("The file exist are different count:%d", exist.Load())
	}
}
