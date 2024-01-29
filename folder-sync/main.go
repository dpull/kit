package main

import (
	"flag"
)

func main() {
	versionFlag := flag.Bool("version", false, "-version dir base_dir output.csv")
	diffFlag := flag.Bool("diff", false, "-diff src_ver.csv dst_ver.csv output.csv")
	syncFlag := flag.Bool("sync", false, "-sync diff.csv src_dir dst_dir")
	checkFlag := flag.Bool("check", false, "-check diff.csv src_dir dst_dir")
	flag.Parse()

	args := flag.Args()
	if *versionFlag {
		if len(args) == 2 {
			folder := args[0]
			output := args[1]
			makeFolderVersion(folder, output).Exec()
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

	if *checkFlag {
		if len(args) == 3 {
			diff := args[0]
			srcDir := args[1]
			dstDir := args[2]

			makeFolderCheck(diff, srcDir, dstDir).Exec()
			return
		}
	}
	flag.Usage()
}
