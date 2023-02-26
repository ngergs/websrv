package filesystem_test

import (
	"github.com/ngergs/websrv/filesystem"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileFs(t *testing.T) {
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	require.Nil(t, err)

	osFs := os.DirFS(testDir)
	readFileFs := &filesystem.ReadFileFS{FS: osFs}
	readFileFsData, err := readFileFs.ReadFile(testFile)
	require.Nil(t, err)

	require.Equal(t, originalData, readFileFsData)
}
