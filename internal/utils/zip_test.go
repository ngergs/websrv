package utils_test

import (
	"compress/gzip"
	"testing"

	"github.com/ngergs/websrv/v3/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestZip(t *testing.T) {
	testMsg := []byte("test123")
	zipped, err := utils.Zip(testMsg, gzip.BestCompression)
	require.NoError(t, err)
	require.NotEqual(t, testMsg, zipped)
	unzipped, err := utils.Unzip(zipped)
	require.NoError(t, err)
	require.Equal(t, testMsg, unzipped)
}

func TestUnzipBadInput(t *testing.T) {
	testMsg := []byte("test123")
	_, err := utils.Unzip(testMsg)
	require.Error(t, err)
}
