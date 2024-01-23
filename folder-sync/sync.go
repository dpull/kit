package main

import (
	"encoding/csv"
	"os"
	"path"
	"sync"
)

const (
	CpProcCoNum = 32
	RmProcCoNum = 2
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

	wg.Add(CpProcCoNum)
	for i := 0; i < CpProcCoNum; i++ {
		go func() {
			defer wg.Done()
			for file := range fs.modified {
				dst := path.Join(fs.dstDir, file)
				dstDir := path.Dir(dst)
				mkdirAll(dstDir)
				copyFile(dst, path.Join(fs.srcDir, file))
			}
		}()
	}

	wg.Add(RmProcCoNum)
	for i := 0; i < RmProcCoNum; i++ {
		go func() {
			defer wg.Done()
			for file := range fs.removed {
				os.Remove(path.Join(fs.dstDir, file))
			}
		}()
	}

	readDiff(fs.diff, fs.modified, fs.removed)
	close(fs.modified)
	close(fs.removed)
	wg.Wait()
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
