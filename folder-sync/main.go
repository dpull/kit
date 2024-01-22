package main

import (
	"flag"
	"io/fs"
	"path/filepath"
	"strings"
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
	versionFlag := flag.Bool("version", false, "-version dir base_dir output.csv")
	syncFlag := flag.Bool("sync", false, "-sync src_ver.csv dst_ver.csv src_dir dst_dir")
	flag.Parse()

	args := flag.Args()
	if *versionFlag {
		if len(args) == 3 {
			folder := args[0]
			basePath := args[1]
			output := args[2]
			makeFolderVersion(folder, basePath, output).Exec()
			return
		}
	}

	if *syncFlag {
		if len(args) == 4 {

			srcVer := args[0]
			dstVer := args[1]
			srcDir := args[2]
			dstDir := args[3]

			makeFolderSync(srcVer, dstVer, srcDir, dstDir).Exec()
			return
		}
	}
	flag.Usage()
}
