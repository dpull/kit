package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
)

type folderSync struct {
	srcVer  string
	dstVer  string
	srcDir  string
	dstDir  string
	changed map[string]fileVersion
	removed map[string]fileVersion
	wg      sync.WaitGroup
}

func makeFolderSync(srcVer, dstVer, srcDir, dstDir string) *folderSync {
	fs := new(folderSync)
	fs.srcVer = srcVer
	fs.dstVer = dstVer
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.changed = make(map[string]fileVersion, 1024*1024)
	return fs
}

func (fs *folderSync) Exec() {
	fs.cmp()
	fs.sync()
}

func (fs *folderSync) cmp() {
	readVerToMap(fs.srcVer, fs.removed)

	filesVer := make(chan fileVersion, 1024)
	go func() {
		defer close(filesVer)
		err := readVersion(fs.dstVer, filesVer)
		if err != nil {
			log.Panic(err)
		}
	}()

	for fileVer := range filesVer {
		src, ok := fs.removed[fileVer.path]
		if ok {
			delete(fs.removed, fileVer.path)
		}
		if ok && src.modTime == fileVer.modTime && src.fileSize == fileVer.fileSize && src.fileCRC == fileVer.fileCRC {
			continue
		}
		fs.changed[fileVer.path] = fileVer
	}
}

func (fs *folderSync) sync() {
	var wg sync.WaitGroup
	wg.Add(len(fs.changed))
	for file := range fs.changed {
		go func(filePath string) {
			defer wg.Done()
			copyFile(path.Join(fs.dstDir, filePath), path.Join(fs.srcDir, filePath))
		}(file)
	}
	for file := range fs.removed {
		os.Remove(file)
	}
	wg.Wait()
}

func readVerToMap(verFile string, verMap map[string]fileVersion) {
	filesVer := make(chan fileVersion, 1024)
	go func() {
		defer close(filesVer)
		err := readVersion(verFile, filesVer)
		if err != nil {
			log.Panic(err)
		}
	}()

	for fileVer := range filesVer {
		verMap[fileVer.path] = fileVer
	}
}

func readVersion(verFile string, filesVer chan<- fileVersion) error {
	file, err := os.Open(verFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	var fileVer fileVersion
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		for i, v := range record {
			switch header[i] {
			case ColPath:
				fileVer.path = v
			case ColModTime:
				num, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				fileVer.modTime = num
			case ColFileSize:
				num, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				fileVer.fileSize = num
			case ColFileCRC:
				num, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return err
				}
				fileVer.fileCRC = num
			}
		}
		filesVer <- fileVer
	}
	return nil
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
