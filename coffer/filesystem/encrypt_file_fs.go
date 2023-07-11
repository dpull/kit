package filesystem

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"golang.org/x/net/webdav"
)

type OpenEncryptFile func(fd *os.File, key []byte) (webdav.File, error)

func CreateEncryptFileFS(folder string, param map[string]string, openFn OpenEncryptFile) (webdav.FileSystem, error) {
	key, exist := param["key"]
	if !exist {
		return nil, errors.Errorf("key not exist:%v", param)
	}
	return &encryptFileFS{
		folder: folder,
		key:    []byte(key),
		openFn: openFn,
	}, nil
}

type encryptFileFS struct {
	folder string
	key    []byte
	openFn OpenEncryptFile
}

func (fs *encryptFileFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name = ResolvePath(fs.folder, name); name == "" {
		return os.ErrNotExist
	}
	return os.Mkdir(name, perm)
}

func (fs *encryptFileFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if name = ResolvePath(fs.folder, name); name == "" {
		return nil, os.ErrNotExist
	}

	fd, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	// log.Printf("open file:%v", name)
	return fs.openFn(fd, fs.key)
}

func (fs *encryptFileFS) RemoveAll(ctx context.Context, name string) error {
	if name = ResolvePath(fs.folder, name); name == "" {
		return os.ErrNotExist
	}
	if name == filepath.Clean(fs.folder) {
		// Prohibit removing the virtual root directory.
		return os.ErrInvalid
	}
	return os.RemoveAll(name)
}

func (fs *encryptFileFS) Rename(ctx context.Context, oldName, newName string) error {
	if oldName = ResolvePath(fs.folder, oldName); oldName == "" {
		return os.ErrNotExist
	}
	if newName = ResolvePath(fs.folder, newName); newName == "" {
		return os.ErrNotExist
	}
	if root := filepath.Clean(fs.folder); root == oldName || root == newName {
		// Prohibit renaming from or to the virtual root directory.
		return os.ErrInvalid
	}
	return os.Rename(oldName, newName)
}

func (fs *encryptFileFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name = ResolvePath(fs.folder, name); name == "" {
		return nil, os.ErrNotExist
	}
	return os.Stat(name)
}
