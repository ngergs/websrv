package filesystem_test

import (
	"compress/gzip"
	"context"
	"github.com/ngergs/websrv/v2/filesystem"
	"io"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/ngergs/websrv/v2/internal/utils"
	"github.com/stretchr/testify/require"
)

const testDir = "../test/benchmark"
const testFile = "dummy_random.js"

// TestMemoryFsReadFile tests is ReadFile from the fs.ReadFileFS interface works
func TestMemoryFsReadFile(t *testing.T) {
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	require.Nil(t, err)
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	require.Nil(t, err)
	memoryData, err := memoryFs.ReadFile(testFile)
	require.Nil(t, err)
	require.Equal(t, originalData, memoryData)
}

// TestMemoryFsOpenFile tests Open from the fs.FS interface
func TestMemoryFsOpenFile(t *testing.T) {
	osFs := os.DirFS(testDir)
	originalData, originalStat := getStatsContent(t, osFs, testFile)
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	require.Nil(t, err)
	memoryData, memoryStat := getStatsContent(t, memoryFs, testFile)

	require.Equal(t, originalStat, memoryStat)
	require.Equal(t, originalData, memoryData)
}

// TestMemoryFsZip tests the zip functionality of the memoryFs
func TestMemoryFsZip(t *testing.T) {
	memoryFs, err := filesystem.NewMemoryFs(testDir)
	require.Nil(t, err)
	memoryFsZipped, err := memoryFs.Zip()
	require.Nil(t, err)

	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	require.Nil(t, err)
	originalDataZipped, err := utils.Zip(originalData, gzip.BestCompression)
	require.Nil(t, err)

	memoryDataZipped, err := memoryFsZipped.ReadFile(testFile)
	require.Nil(t, err)
	require.Equal(t, originalDataZipped, memoryDataZipped)
}

func getStatsContent(t *testing.T, fs fs.FS, path string) ([]byte, fs.FileInfo) {
	file, err := fs.Open(path)
	require.Nil(t, err)
	defer utils.Close(context.Background(), file)
	stat, err := file.Stat()
	require.Nil(t, err)
	data, err := io.ReadAll(file)
	require.Nil(t, err)
	return data, stat
}
