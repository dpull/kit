package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/bytedance/gopkg/util/gopool"
)

func copyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func mkdirAll(dir string) error {
	if !exists(dir) {
		err := os.MkdirAll(dir, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func WalkDir(dir string, files chan<- string) {
	var wg sync.WaitGroup
	wg.Add(1)
	gopool.Go(func() {
		walk(&wg, files, dir)
	})
	wg.Wait()
}

func walk(wg *sync.WaitGroup, files chan<- string, dir string) {
	defer wg.Done()

	entry, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("readdir %s failed, %s", dir, err)
		return
	}
	for _, e := range entry {
		fullPath := filepath.Join(dir, e.Name())
		if e.IsDir() {
			wg.Add(1)
			gopool.Go(func() {
				walk(wg, files, fullPath)
			})
		} else {
			files <- fullPath
		}
	}
}
