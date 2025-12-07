package filesystem_test

import (
	"github.com/ngergs/websrv/v4/filesystem"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileFs(t *testing.T) {
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	require.NoError(t, err)

	osFs := os.DirFS(testDir)
	readFileFs := &filesystem.ReadFileFS{FS: osFs}
	readFileFsData, err := readFileFs.ReadFile(testFile)
	require.NoError(t, err)

	require.Equal(t, originalData, readFileFsData)
}
