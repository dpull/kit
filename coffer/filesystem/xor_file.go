package filesystem

import (
	"io"
	"os"

	"golang.org/x/net/webdav"
)

func init() {
	Register("xor", func(folder string, param map[string]string) (webdav.FileSystem, error) {
		return CreateEncryptFileFS(folder, param,
			func(fd *os.File, key []byte) (webdav.File, error) {
				stat, _ := fd.Stat()
				if stat.IsDir() {
					return fd, nil
				}
				return &xorFile{fd: fd, key: key}, nil
			})
	})
}

type xorFile struct {
	fd  *os.File
	key []byte
}

func xor(data []byte, offset int64, key []byte) {
	dataLen := int64(len(data))
	keyLen := int64(len(key))
	for i := int64(0); i < dataLen; i++ {
		seed := key[(offset+i)%keyLen]
		data[i] ^= seed
	}
}

func (f *xorFile) Read(b []byte) (int, error) {
	pos, err := f.fd.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	n, err := f.fd.Read(b)
	if n <= 0 {
		return n, err
	}
	xor(b[0:n], pos, f.key)
	return n, err
}

func (f *xorFile) Write(b []byte) (int, error) {
	pos, err := f.fd.Seek(0, 1)
	if err != nil {
		return 0, err
	}

	data := make([]byte, len(b))
	copy(data, b)
	xor(data, pos, f.key)
	return f.fd.Write(data)
}

func (f *xorFile) Close() error {
	return f.fd.Close()
}

func (f *xorFile) Readdir(n int) ([]os.FileInfo, error) {
	return f.fd.Readdir(n)
}

func (f *xorFile) Seek(offset int64, whence int) (int64, error) {
	return f.fd.Seek(offset, whence)
}

func (f *xorFile) Stat() (os.FileInfo, error) {
	return f.fd.Stat()
}
