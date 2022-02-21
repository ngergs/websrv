package filesystem_test

import (
	"os"
	"path"
	"testing"

	"github.com/ngergs/webserver/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestReadFileFs(t *testing.T) {
	originalData, err := os.ReadFile(path.Join(testDir, testFile))
	assert.Nil(t, err)

	osFs := os.DirFS(testDir)
	readFileFs := &filesystem.ReadFileFS{osFs}
	readFileFsData, err := readFileFs.ReadFile(testFile)
	assert.Nil(t, err)

	assert.Equal(t, originalData, readFileFsData)
}
