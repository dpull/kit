package filesystem

import (
	"github.com/pkg/errors"
	"golang.org/x/net/webdav"
)

type createFn func(folder string, param map[string]string) (webdav.FileSystem, error)

var factory = make(map[string]createFn)

func Register(fsType string, create createFn) {
	factory[fsType] = create
}

func Create(fsType, folder string, param map[string]string) (webdav.FileSystem, error) {
	create, exist := factory[fsType]
	if !exist {
		return nil, errors.Errorf("fsType not exist:%s", fsType)
	}
	return create(folder, param)
}
