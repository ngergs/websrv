package filesystem

import (
	"io"
	"io/fs"
)

// MemoryFilesystem only holds actual files, not the directory entries
type ReadFileFS struct {
	fs.FS
}

func (fs *ReadFileFS) ReadFile(name string) ([]byte, error) {
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
