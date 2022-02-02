package filesystem

import (
	"io"
	"io/fs"
)

type ZipFs interface {
	fs.ReadFileFS
	IsZipped(path string) (bool, error)
}

type unzippedFs struct {
	filesystem fs.FS
}

func (fs *unzippedFs) Open(name string) (fs.File, error) {
	return fs.filesystem.Open(name)
}

func (fs *unzippedFs) IsZipped(path string) (bool, error) {
	return false, nil
}

func (fs *unzippedFs) ReadFile(name string) ([]byte, error) {
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

func FromUnzippedFs(filesystem fs.FS) ZipFs {
	return &unzippedFs{filesystem: filesystem}
}
