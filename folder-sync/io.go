package main

import (
	"bufio"
	"bytes"
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

func cmpFile(file1, file2 string) (bool, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return false, err
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return false, err
	}

	if len(data1) != len(data2) {
		return false, nil
	}
	return bytes.Compare(data1, data2) == 0, nil
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

func init() {
	gopool.SetCap(1024)
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

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	r := bufio.NewReader(f)
	for {
		// ReadLine is a low-level line-reading primitive.
		// Most callers should use ReadBytes('\n') or ReadString('\n') instead or use a Scanner.
		bytes, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return lines, err
		}
		lines = append(lines, string(bytes))
	}
	return lines, nil
}
