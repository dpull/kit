package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func ignoreName(path string) bool {
	return strings.Contains(path, ".svn") || strings.Contains(path, ".git")
}

func findAllPaths(dir string, paths chan<- string) {
	if ignoreName(dir) {
		return
	}

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if dir == path {
			return nil
		}

		paths <- path

		if d.IsDir() {
			go findAllPaths(path, paths)
			return nil
		}
		return nil
	})
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage:disable-readonly dir")
		return
	}
	dir := os.Args[1]
	paths := make(chan string, 100)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for path := range paths {
				st, err := os.Stat(path)
				if err != nil {
					fmt.Println(path, err)
					continue
				}
				err = os.Chmod(path, st.Mode()&^0400)
				if err != nil {
					fmt.Println(path, err)
					continue
				}
			}
		}()
	}

	findAllPaths(dir, paths)
	close(paths)
	wg.Wait()
}
