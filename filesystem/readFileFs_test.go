package filesystem_test

import (
	"github.com/ngergs/websrv/filesystem"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFileFs(t *testing.T) {
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	assert.Nil(t, err)

	osFs := os.DirFS(testDir)
	readFileFs := &filesystem.ReadFileFS{FS: osFs}
	readFileFsData, err := readFileFs.ReadFile(testFile)
	assert.Nil(t, err)

	assert.Equal(t, originalData, readFileFsData)
}
