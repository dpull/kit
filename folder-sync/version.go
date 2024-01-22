package main

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	FileProcCoNum = 2048

	ColPath     = "Path"
	ColModTime  = "ModTime"
	ColFileSize = "FileSize"
	ColFileCRC  = "FileCRC"
)

type folerVersion struct {
	wd       walkDir
	filesVer chan fileVersion
	wg       sync.WaitGroup
	folder   string
	basePath string
	output   string
}

func makeFolderVersion(folder, basePath, output string) *folerVersion {
	fv := new(folerVersion)
	fv.filesVer = make(chan fileVersion, 1024)
	fv.folder = folder
	fv.basePath = basePath
	fv.output = output
	fv.wd.Init()
	return fv
}

func (fv *folerVersion) Exec() {
	fv.wg.Add(3)

	go func() {
		defer fv.wg.Done()
		fv.wd.Exec(fv.folder)
	}()

	go func() {
		defer fv.wg.Done()
		defer close(fv.filesVer)
		getFilesVersion(fv.basePath, fv.wd.Files, fv.filesVer)
	}()

	go func() {
		defer fv.wg.Done()
		err := outputVersion(fv.output, fv.filesVer)
		if err != nil {
			log.Panicf("get all files failed, %s", err)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			log.Printf("[stat]files:%d, filesVer:%d", len(fv.wd.Files), len(fv.filesVer))
		}
	}()

	fv.wg.Wait()
}

func getFilesVersion(basePath string, files chan string, filesVer chan<- fileVersion) {
	var wg sync.WaitGroup
	wg.Add(FileProcCoNum)

	for i := 0; i < FileProcCoNum; i++ {
		go func() {
			defer wg.Done()
			var fileVer fileVersion

			for path := range files {
				st, err := os.Stat(path)
				if err != nil {
					log.Printf("get path:%s stat failed, %s", path, err)
					continue
				}

				fileVer.path, err = filepath.Rel(basePath, path)
				if err != nil {
					log.Printf("get path:%s|%s rel failed, %s", basePath, path, err)
					continue
				}

				fileVer.modTime = st.ModTime().Unix()
				fileVer.fileSize = st.Size()
				fileVer.fileCRC = 0
				filesVer <- fileVer
			}
		}()
	}
	wg.Wait()
}

type fileVersion struct {
	path     string
	modTime  int64
	fileSize int64
	fileCRC  uint64
}

func outputVersion(output string, filesVer chan fileVersion) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write([]string{ColPath, ColModTime, ColFileSize, ColFileCRC})
	if err != nil {
		return err
	}
	row := make([]string, 4)
	for fileVer := range filesVer {
		row[0] = fileVer.path
		row[1] = strconv.FormatInt(fileVer.modTime, 10)
		row[2] = strconv.FormatInt(fileVer.fileSize, 10)
		row[3] = strconv.FormatUint(fileVer.fileCRC, 10)

		err = writer.Write(row)
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}
