package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"
	"time"
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
	traceFlag := flag.Bool("trace", false, "-trace")
	versionFlag := flag.Bool("version", false, "-version dir output.csv")
	syncFlag := flag.Bool("sync", false, "-sync src_ver.csv dst_ver.csv src_dir dst_dir")
	flag.Parse()

	args := flag.Args()
	if *traceFlag {
		fileName := fmt.Sprintf("file-%s.trace", time.Now().Format("2006-01-02-15-04-05"))
		file, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		err = trace.Start(file)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	if *versionFlag {
		if len(args) == 2 {
			folder := args[0]
			output := args[1]
			makeFolderVersion(folder, output).Exec()
			return
		}
	}

	if *syncFlag {
		if len(args) == 4 {
			/*
				srcVersion := args[0]
				dstVersion := args[1]
				srcDir := args[2]
				dstDir := args[3]
			*/
			return
		}
	}
	flag.Usage()
}
