package filesystems

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// MemoryFilesystem only holds actual files, not the directory entries
type MemoryFilesystem struct {
	files           map[string]*memoryFile
	targetDirLength int
}

type memoryFile struct {
	data      []byte
	info      fs.FileInfo
	dirInfo   []fs.DirEntry
	dirOffset int
}

type openMemoryFile struct {
	readOffset int
	file       *memoryFile
}

func NewMemoryFs(targetDir string) (*MemoryFilesystem, error) {
	targetDir = path.Clean(targetDir)
	files := make(map[string]*memoryFile)
	fs := &MemoryFilesystem{
		files:           files,
		targetDirLength: len(targetDir),
	}
	err := filepath.Walk(targetDir, fs.readFile)
	if err != nil {
		return nil, fmt.Errorf("error reading files into in-memory-fs: %w", err)
	}
	return fs, nil
}

func (fs *MemoryFilesystem) readFile(path string, info fs.FileInfo, err error) error {
	// remove targetDir part and leading / from path
	var subPath string
	if len(path) > fs.targetDirLength {
		subPath = path[(fs.targetDirLength + 1):]
	} else {
		subPath = "."
	}
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var result *memoryFile
	if info.IsDir() {
		dirInfo, err := file.ReadDir(0)
		if err != nil {
			return err
		}
		result = &memoryFile{info: info, dirInfo: dirInfo}
	} else {
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		result = &memoryFile{data: data, info: info}
	}
	log.Debug().Msgf("Read into memory-fs: %s", subPath)
	fs.files[subPath] = result
	return nil
}

func (fs *MemoryFilesystem) Open(name string) (fs.File, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found", name)
	}
	return &openMemoryFile{file: file}, nil
}

func (file *memoryFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if !file.info.IsDir() {
		return make([]fs.DirEntry, 0), fmt.Errorf("not a directory")
	}
	if n <= 0 {
		return file.dirInfo, nil
	}
	if len(file.dirInfo) <= file.dirOffset {
		return make([]fs.DirEntry, 0), io.EOF
	}
	if len(file.dirInfo) < file.dirOffset+n {
		n = len(file.dirInfo) - file.dirOffset
	}
	fmt.Printf("%d\n", n)
	result := file.dirInfo[file.dirOffset : file.dirOffset+n]
	file.dirOffset += n
	return result, nil
}

func (open *openMemoryFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return open.file.ReadDir(n)
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

func (open *openMemoryFile) WriteTo(w io.Writer) (int64, error) {
	if open.readOffset >= len(open.file.data) {
		return 0, io.EOF
	}
	for open.readOffset < len(open.file.data) {
		n, err := w.Write(open.file.data[open.readOffset:])
		open.readOffset += n
		if err != nil {
			return int64(open.readOffset), err
		}
	}
	return int64(open.readOffset), nil
}

func (file *openMemoryFile) Close() error {
	// in memory file does nothing on error
	return nil
}
