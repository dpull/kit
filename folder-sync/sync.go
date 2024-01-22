package main

import (
	"encoding/csv"
	"io"
	"os"
	"path"
	"sync"
)

const (
	CopyProcCoNum = 1024
	FileProcCoNum = 2048
)

type folderSync struct {
	diff     string
	srcDir   string
	dstDir   string
	removed  chan string
	modified chan string
}

func makeFolderSync(diff, srcDir, dstDir string) *folderSync {
	fs := new(folderSync)
	fs.diff = diff
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.removed = make(chan string, 64)
	fs.modified = make(chan string, 1024)
	return fs
}

func (fs *folderSync) Exec() {
	var wg sync.WaitGroup
	wg.Add(16)
	for i := 0; i < 16; i++ {
		go func() {
			defer wg.Done()
			for file := range fs.modified {
				copyFile(path.Join(fs.dstDir, file), path.Join(fs.srcDir, file))
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

	}()
	for file := range fs.changed {
		go func(filePath string) {
			copyFile(path.Join(fs.dstDir, filePath), path.Join(fs.srcDir, filePath))
		}(file)
	}
	for file := range fs.removed {
		os.Remove(path.Join(fs.dstDir, file))
	}
	wg.Wait()
}

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

func readDiff(diffFile string, modified, removed chan<- string) error {
	file, err := os.Open(diffFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		switch record[0] {
		case OpMod:
			modified <- record[1]
		case OpDel:
			removed <- record[1]
		}
	}
	return nil
}
