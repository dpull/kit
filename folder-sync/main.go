package main

import (
	"flag"
)

func main() {
	versionFlag := flag.Bool("version", false, "-version dir base_dir output.csv")
	diffFlag := flag.Bool("compare", false, "-diff src_ver.csv dst_ver.csv output.csv")
	syncFlag := flag.Bool("sync", false, "-sync diff.csv src_dir dst_dir")
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

	if *diffFlag {
		if len(args) == 3 {

			srcVer := args[0]
			dstVer := args[1]
			output := args[2]

			makeVersionDiff(srcVer, dstVer, output).Exec()
			return
		}
	}

	if *syncFlag {
		if len(args) == 3 {
			diff := args[0]
			srcDir := args[1]
			dstDir := args[2]

			makeFolderSync(diff, srcDir, dstDir).Exec()
			return
		}
	}
	flag.Usage()
}
