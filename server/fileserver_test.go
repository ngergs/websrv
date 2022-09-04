package server_test

import (
	"compress/gzip"
	"github.com/ngergs/websrv/server"
	"net/http"
	"net/url"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/assert"
)

const testDir = "../test/benchmark"
const testFile = "dummy_random.js"
const fallbackFile = "index.html"

// TestFileServerSimpleServe check sif a plain file without any extras is delivered
func TestFileServerSimpleServe(t *testing.T) {
	w, _, r := getHandlerMockWithPath(t, testFile)
	handler, originalFileData, _ := getWebserverHandler(t, []string{})
	handler.ServeHTTP(w, r)
	assert.Equal(t, originalFileData, w.receivedData.Bytes())
}

// TestWebServerSimpleServe check sif a plain file without any extras is delivered
func TestFileServerFallback(t *testing.T) {
	w, _, r := getHandlerMockWithPath(t, "non-existing")
	handler, _, originalFallbackData := getWebserverHandler(t, []string{})
	handler.ServeHTTP(w, r)
	assert.Equal(t, originalFallbackData, w.receivedData.Bytes())
}

// TestWebServerSimpleServe check sif a plain file without any extras is delivered
func TestFileServerZip(t *testing.T) {
	w, responseHeader, r := getHandlerMockWithPath(t, testFile)
	r.Header.Set("Accept-Encoding", "gzip")
	handler, originalFileData, _ := getWebserverHandler(t, []string{"application/javascript"})
	originalFileDataZipped, err := utils.Zip(originalFileData, gzip.BestCompression)
	assert.Nil(t, err)
	handler.ServeHTTP(w, r)
	assert.Equal(t, originalFileDataZipped, w.receivedData.Bytes())
	assert.Equal(t, "gzip", responseHeader.Get("Content-Encoding"))
}

func TestFileServerZipFallback(t *testing.T) {
	w, responseHeader, r := getHandlerMockWithPath(t, "non-existing")
	r.Header.Set("Accept-Encoding", "gzip")
	handler, _, originalFallbackData := getWebserverHandler(t, []string{"text/html"})
	originalFileDataZipped, err := utils.Zip(originalFallbackData, gzip.BestCompression)
	assert.Nil(t, err)
	handler.ServeHTTP(w, r)
	assert.Equal(t, originalFileDataZipped, w.receivedData.Bytes())
	assert.Equal(t, "gzip", responseHeader.Get("Content-Encoding"))
}

func getWebserverHandler(t *testing.T, zipMediaTypes []string) (handler http.Handler, originalData []byte, fallbackData []byte) {
	fs, err := filesystem.NewMemoryFs(testDir)
	assert.Nil(t, err)
	zippedFs, err := fs.Zip([]string{".html", ".js"})
	assert.Nil(t, err)
	originalData, err = fs.ReadFile(testFile)
	assert.Nil(t, err)
	fallbackData, err = fs.ReadFile(fallbackFile)
	assert.Nil(t, err)
	return server.FileServerHandler(fs, zippedFs, fallbackFile, &server.Config{
		MediaTypeMap:   map[string]string{".html": "text/html", ".js": "application/javascript"},
		GzipMediaTypes: zipMediaTypes,
	}), originalData, fallbackData
}

func getHandlerMockWithPath(t *testing.T, path string) (responseWriter *mockResponseWriter, responseHeader *http.Header, request *http.Request) {
	w, r, _ := getDefaultHandlerMocks()
	url, err := url.Parse(path)
	assert.Nil(t, err)
	r.URL = url
	rHeader := http.Header(make(map[string][]string))
	w.mock.On("Header").Return(rHeader)
	return w, &rHeader, r
}
