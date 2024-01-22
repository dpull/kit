package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"
)

const (
	OpMod = "M"
	OpDel = "D"
)

type folderDiff struct {
	srcVer   string
	dstVer   string
	output   string
	modified map[string]fileVersion
	removed  map[string]fileVersion
	wg       sync.WaitGroup
}

func makeVersionDiff(srcVer, dstVer, output string) *folderDiff {
	fd := new(folderDiff)
	fd.srcVer = srcVer
	fd.dstVer = dstVer
	fd.output = output
	fd.modified = make(map[string]fileVersion, 1024*1024)
	fd.removed = make(map[string]fileVersion, 1024*1024)
	return fd
}

func (fd *folderDiff) Exec() {
	readVerToMap(fd.srcVer, fd.removed)

	filesVer := make(chan fileVersion, 1024)
	go func() {
		defer close(filesVer)
		err := readVersion(fd.dstVer, filesVer)
		if err != nil {
			log.Panic(err)
		}
	}()

	for fileVer := range filesVer {
		src, ok := fd.removed[fileVer.path]
		if ok {
			delete(fd.removed, fileVer.path)
		}
		if ok && src.modTime == fileVer.modTime && src.fileSize == fileVer.fileSize && src.fileCRC == fileVer.fileCRC {
			continue
		}
		fd.modified[fileVer.path] = fileVer
	}
	err := outputDiff(fd.output, fd.modified, fd.removed)
	if err != nil {
		log.Panic(err)
	}
}

func outputDiff(output string, modified, removed map[string]fileVersion) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	row := make([]string, 2)
	for file := range modified {
		row[0] = OpMod
		row[1] = file
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}
	for file := range removed {
		row[0] = OpDel
		row[1] = file
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
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
