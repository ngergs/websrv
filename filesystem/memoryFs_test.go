package filesystem_test

import (
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/utils"
	"github.com/stretchr/testify/assert"
)

const testDir = "../benchmark"
const testFile = "dummy_random.js"

// TestMemoryFsReadFile tests is ReadFile from the fs.ReadFileFS interface works
func TestMemoryFsReadFile(t *testing.T) {
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	assert.Nil(t, err)
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	assert.Nil(t, err)
	memoryData, err := memoryFs.ReadFile(testFile)
	assert.Nil(t, err)
	assert.Equal(t, originalData, memoryData)
}

// TestMemoryFsOpenFile tests Open from the fs.FS interface
func TestMemoryFsOpenFile(t *testing.T) {
	osFs := os.DirFS(testDir)
	originalData, originalStat := getStatsContent(t, osFs, testFile)
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	assert.Nil(t, err)
	memoryData, memoryStat := getStatsContent(t, memoryFs, testFile)

	assert.Equal(t, originalStat, memoryStat)
	assert.Equal(t, originalData, memoryData)
}

// TestMemoryFsZip tests the zip functionality of the memoryFs
func TestMemoryFsZip(t *testing.T) {
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	assert.Nil(t, err)
	memoryFsZipped, err := memoryFs.Zip([]string{".js"})
	assert.Nil(t, err)

	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	assert.Nil(t, err)
	originalDataZipped, err := utils.Zip(originalData, gzip.BestCompression)
	assert.Nil(t, err)

	memoryDataZipped, err := memoryFsZipped.ReadFile(testFile)
	assert.Nil(t, err)
	assert.Equal(t, originalDataZipped, memoryDataZipped)
}

// TestMemoryFsZipNonMatch tests that file sthat do not match the zip file extension are not present in the zipped memoryFs.
func TestMemoryFsZipNonMatch(t *testing.T) {
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	assert.Nil(t, err)
	memoryFsZipped, err := memoryFs.Zip([]string{})
	assert.Nil(t, err)
	_, err = memoryFsZipped.ReadFile(testFile)
	// file extension does not match. Hence, the given testFile is not present in the zipped memoryFs
	assert.NotNil(t, err)
}

func getStatsContent(t *testing.T, fs fs.FS, path string) ([]byte, fs.FileInfo) {
	file, err := fs.Open(path)
	assert.Nil(t, err)
	defer file.Close()
	stat, err := file.Stat()
	assert.Nil(t, err)
	data, err := io.ReadAll(file)
	assert.Nil(t, err)
	return data, stat
}
