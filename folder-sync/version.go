package main

import (
	"encoding/csv"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type folerVersion struct {
	files    chan string
	filesVer chan fileVersion
	wg       sync.WaitGroup
	folder   string
	output   string
}

func makeFolderVersion(folder, output string) *folerVersion {
	fv := new(folerVersion)
	fv.files = make(chan string, 128)
	fv.filesVer = make(chan fileVersion, 32)
	fv.folder = folder
	fv.output = output

	return fv
}

func (fv *folerVersion) Exec() {
	fv.wg.Add(3)

	go func() {
		defer fv.wg.Done()
		getAllFiles(fv.folder, fv.files)
		close(fv.files)
	}()

	go func() {
		defer fv.wg.Done()
		getFilesVersion(fv.files, fv.filesVer)
		close(fv.filesVer)
	}()

	go func() {
		defer fv.wg.Done()
		outputVersion(fv.output, fv.filesVer)
	}()

	fv.wg.Wait()
}

func getFilesVersion(files chan string, filesVer chan<- fileVersion) {
	var wg sync.WaitGroup
	wg.Add(128)

	for i := 0; i < 128; i++ {
		go func() {
			defer wg.Done()
			var fileVer fileVersion

			for path := range files {
				st, err := os.Stat(path)
				if err != nil {
					log.Printf("get path:%s stat failed, %s", path, err)
					continue
				}

				fileVer.path = path
				fileVer.modTime = st.ModTime().Unix()
				fileVer.fileSize = st.Size()
				fileVer.fileCRC = 0
				filesVer <- fileVer
			}
		}()
	}
	wg.Wait()
}

func getAllFiles(dir string, files chan<- string) error {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walk dir:%s, path:%s failed, %s", dir, path, err)
			return nil
		}

		if dir == path {
			return nil
		}

		if d.IsDir() {
			return getAllFiles(path, files)
		}

		files <- path
		return nil
	})
	return err
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
	err = writer.Write([]string{"Path", "ModTime", "FileSize", "FileCRC"})
	if err != nil {
		return err
	}
	row := make([]string, 4)
	for fileVer := range filesVer {
		row[0] = fileVer.path
		row[1] = strconv.FormatInt(fileVer.modTime, 10)
		row[2] = strconv.FormatInt(fileVer.fileSize, 10)
		row[3] = strconv.FormatUint(fileVer.fileCRC, 10)

		err = writer.Write([]string{"Path", "ModTime", "FileSize", "FileCRC"})
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}
