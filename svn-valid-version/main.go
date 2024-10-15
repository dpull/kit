package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bytedance/gopkg/util/gopool"
)

func ignoreName(path string) bool {
	return strings.Contains(path, ".svn") || strings.Contains(path, ".git")
}

func findAllPaths(dir string, paths chan<- string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if ignoreName(path) {
			return nil
		}
		paths <- path
		return nil
	})
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage:svn-valid-version outfile")
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	outfile, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	var wg sync.WaitGroup

	versions := validVersion{
		versions: map[string]map[string]bool{},
	}

	paths := make(chan string, 100)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for path := range paths {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				fmt.Println(err)
				continue
			}

			wg.Add(1)
			gopool.Go(func() {
				defer wg.Done()

				err := proc(rel, &versions)
				if err != nil {
					log.Printf("%s failed: %+v", rel, err)
				}
			})
		}
	}()

	findAllPaths(dir, paths)
	close(paths)
	wg.Wait()

	err = versions.output(outfile)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}

func proc(file string, versions *validVersion) error {
	svn := SVN{}

	blame, err := svn.Blame(file)
	if err != nil {
		return err
	}

	ver, err := getVersion(blame)
	if err != nil {
		return err
	}

	filterEmptyLine(ver, file)

	versions.add(file, ver)
	return nil
}
