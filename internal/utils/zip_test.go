package utils_test

import (
	"compress/gzip"
	"testing"

	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	testMsg := []byte("test123")
	zipped, err := utils.Zip(testMsg, gzip.BestCompression)
	assert.Nil(t, err)
	assert.NotEqual(t, testMsg, zipped)
	unzipped, err := utils.Unzip(zipped)
	assert.Nil(t, err)
	assert.Equal(t, testMsg, unzipped)
}

func TestUnzipBadInput(t *testing.T) {
	testMsg := []byte("test123")
	_, err := utils.Unzip(testMsg)
	assert.NotNil(t, err)
}
