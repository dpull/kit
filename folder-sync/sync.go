package main

type folderSync struct {
	srcVer  string
	dstVer  string
	srcDir  string
	dstDir  string
	rmFiles map[string]folerVersion
}

func makeFolderSync(srcVer, dstVer, srcDir, dstDir string) *folderSync {
	fs := new(folderSync)
	fs.srcVer = srcVer
	fs.dstVer = dstVer
	fs.srcDir = srcDir
	fs.dstDir = dstDir
	fs.rmFiles = make(map[string]folerVersion, 1024*1024)
	return fs
}
