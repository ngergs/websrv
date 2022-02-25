package filesystem

import (
	"io"
	"io/fs"
)

// ReadFileFS wraps a fs.FS and adds the ReadFile method
type ReadFileFS struct {
	fs.FS
}

// ReadFile is a more concise way to directly read a file into memory.
func (fs *ReadFileFS) ReadFile(name string) ([]byte, error) {
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
