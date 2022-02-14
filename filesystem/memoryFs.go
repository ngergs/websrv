package filesystem

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/ngergs/webserver/utils"
	"github.com/rs/zerolog/log"
)

// MemoryFS only holds actual files, not the directory entries
type MemoryFS struct {
	files map[string]*memoryFile
}

type memoryFile struct {
	data    []byte
	info    fs.FileInfo
	dirInfo []fs.DirEntry
}

type openMemoryFile struct {
	readOffset int
	dirOffset  int
	file       *memoryFile
}

func NewMemoryFs(targetDir string) (*MemoryFS, error) {
	targetDir = path.Clean(targetDir)
	fs := &MemoryFS{
		files: make(map[string]*memoryFile),
	}
	err := filepath.Walk(targetDir, getReadFileFunc(fs, len(targetDir)))
	if err != nil {
		return nil, fmt.Errorf("error reading files into in-memory-fs: %w", err)
	}
	return fs, nil
}

func getReadFileFunc(filesystem *MemoryFS, targetDirLength int) func(path string, info fs.FileInfo, err error) error {
	return func(filePath string, info fs.FileInfo, err error) error {
		// remove targetDir part and leading / from path
		var subPath string
		if len(filePath) > targetDirLength {
			subPath = filePath[(targetDirLength + 1):]
		} else {
			subPath = "."
		}
		if err != nil {
			return err
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer utils.Close(context.Background(), file)

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
		filesystem.files[subPath] = result
		return nil
	}
}

func (fs *MemoryFS) Open(name string) (fs.File, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found", name)
	}
	return &openMemoryFile{file: file}, nil
}
func (fs *MemoryFS) ReadFile(name string) ([]byte, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found in filesystem", name)
	}
	return file.data, nil
}

type modifiedSizeInfo struct {
	size int64
	fs.FileInfo
}

func (mod *modifiedSizeInfo) Size() int64 {
	return mod.size
}

// Zip returns a deep copy of the filesystem where all files that match the given zip file extension are zipped.
// Files that do not match are absent in the zipped version of the in memoryfilesystem.
func (fs *MemoryFS) Zip(zipFileExtensions []string) (*MemoryFS, error) {
	zippedFiles := make(map[string]*memoryFile)
	for filepath, file := range fs.files {
		if utils.Contains(zipFileExtensions, path.Ext(filepath)) {
			log.Debug().Msgf("Zipping %s", filepath)
			zipped, err := utils.Zip(file.data, gzip.BestCompression)
			if err != nil {
				return nil, err
			}
			info := &modifiedSizeInfo{size: int64(len(zipped)), FileInfo: file.info}
			zippedFiles[filepath] = &memoryFile{data: zipped, info: info}
		}
	}
	return &MemoryFS{files: zippedFiles}, nil
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
func (open *openMemoryFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if open.dirOffset >= len(open.file.dirInfo) {
		return []fs.DirEntry{}, io.EOF
	}
	if n <= 0 {
		n = len(open.file.dirInfo) - open.dirOffset
	}
	if n > len(open.file.dirInfo)-open.dirOffset {
		n = len(open.file.dirInfo) - open.dirOffset
	}
	result := open.file.dirInfo[open.dirOffset : open.dirOffset+n]
	open.dirOffset += n
	return result, nil
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
