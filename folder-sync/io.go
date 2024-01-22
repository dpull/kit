package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
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

type walkDir struct {
	wg    sync.WaitGroup
	Files chan string
}

func (wd *walkDir) walk(dir string) {
	defer wd.wg.Done()

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("readdir %s failed, %s", dir, err)
		return
	}
	for _, file := range files {
		fullPath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			wd.wg.Add(1)
			go wd.walk(fullPath)
		} else {
			wd.Files <- fullPath
		}
	}
}

func (wd *walkDir) Init() {
	wd.Files = make(chan string, 1024*1024)
}

func (wd *walkDir) Exec(dir string) {
	wd.wg.Add(1)
	go wd.walk(dir)
	wd.wg.Wait()
	close(wd.Files)
}
