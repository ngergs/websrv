package memoryfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

//MemoryFilesystem only holds actual files, directory entries are discarded
type MemoryFilesystem struct {
	files map[string]*memoryFile
}

type memoryFile struct {
	data []byte
	info fs.FileInfo
}

type openMemoryFile struct {
	readOffset int
	file       *memoryFile
}

func New(targetDir string) *MemoryFilesystem {
	targetDir = path.Clean(targetDir)
	files := make(map[string]*memoryFile)
	filepath.Walk(targetDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		// remove targetDir part and leading / from path
		log.Debug().Msgf("Read into memory-fs: %s", path[(len(targetDir)+1):])
		files[path[(len(targetDir)+1):]] = &memoryFile{data: data, info: info}
		return nil
	})
	return &MemoryFilesystem{
		files: files,
	}
}

func (fs *MemoryFilesystem) Open(name string) (fs.File, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found", name)
	}
	return &openMemoryFile{file: file}, nil
}

func (open *openMemoryFile) Stat() (fs.FileInfo, error) {
	return open.file.info, nil
}
func (open *openMemoryFile) Read(dst []byte) (int, error) {
	if open.readOffset >= len(open.file.data) {
		return 0, io.EOF
	}
	n := copy(dst, open.file.data[open.readOffset:])
	open.readOffset += n
	return n, nil
}

func (file *openMemoryFile) Close() error {
	// in memory file does nothing on error
	return nil
}
