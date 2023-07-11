package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

func ignoreFileName(path string) bool {
	name := filepath.Base(path)
	switch name {
	case ".WeDrive":
		return true
	}
	return false
}

func findEmptyDir(dir string, empty chan<- string, nonEmpth chan<- string) (int, error) {
	count := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if dir == path {
			return nil
		}

		if d.IsDir() {
			subCount, _ := findEmptyDir(path, empty, nonEmpth)
			count += subCount
			return nil
		}

		if !ignoreFileName(path) {
			count++
		}
		return nil
	})
	if count == 0 {
		empty <- dir
	} else {
		nonEmpth <- dir
	}
	return count, err
}

func hideFile(path string) error {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return err
	}

	attrs |= syscall.FILE_ATTRIBUTE_HIDDEN
	return syscall.SetFileAttributes(p, attrs)
}

func showFile(path string) error {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return err
	}

	attrs &^= syscall.FILE_ATTRIBUTE_HIDDEN
	return syscall.SetFileAttributes(p, attrs)
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage:hidden-empty-dir dir")
		return
	}
	dir := os.Args[1]
	emptyDir := make(chan string, 100)
	nonEmptyDir := make(chan string, 100)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for dir := range emptyDir {
			hideFile(dir)
		}
	}()
	go func() {
		defer wg.Done()
		for dir := range nonEmptyDir {
			showFile(dir)
		}
	}()
	findEmptyDir(dir, emptyDir, nonEmptyDir)
	close(emptyDir)
	close(nonEmptyDir)
	wg.Wait()
}
