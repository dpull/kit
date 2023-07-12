package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

func ignoreName(path string) bool {
	name := filepath.Base(path)
	switch name {
	case ".svn":
		return true
	}
	return false
}

func findEmptyDir(dir string, empty chan<- string) (int, error) {
	count := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if dir == path {
			return nil
		}

		if ignoreName(path) {
			return nil
		}

		if d.IsDir() {
			subCount, _ := findEmptyDir(path, empty)
			count += subCount
			return nil
		}

		return nil
	})
	if count == 0 {
		empty <- dir
	}
	return count, err
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage:remove-empty-dir dir")
		return
	}
	dir := os.Args[1]
	emptyDir := make(chan string, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for dir := range emptyDir {
			err := os.Remove(dir)
			if err != nil {
				fmt.Printf("remove-empty-dir %s failed:%s\n", dir, err)
				continue
			}
			fmt.Printf("remove-empty-dir %s\n", dir)
		}
	}()

	findEmptyDir(dir, emptyDir)
	close(emptyDir)
	wg.Wait()
}
