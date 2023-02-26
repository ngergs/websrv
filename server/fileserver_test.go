package server_test

import (
	"compress/gzip"
	"github.com/ngergs/websrv/server"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/assert"
)

const testDir = "../test/benchmark"
const testFile = "dummy_random.js"
const empptyDir = "emptyDir"
const fallbackFile = "index.html"

// TestFileServerSimpleServe checks if a plain file without any extras is delivered
func TestFileServerSimpleServe(t *testing.T) {
	w, r := getHandlerMockWithPath(t, testFile)
	handler, originalFileData, _ := getWebserverHandler(t, []string{})
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, originalFileData, getReceivedData(t, w.Result().Body))
}

// TestDirFallback checks that requesting a directory instead of a file returns the fallbackfile
func TestDirFallback(t *testing.T) {
	w, r := getHandlerMockWithPath(t, empptyDir)
	handler, _, fallbackData := getWebserverHandler(t, []string{})
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, fallbackData, getReceivedData(t, w.Result().Body))
}

// TestWebServerSimpleServe checks that non existing path yields a fallback
func TestFileServerFallback(t *testing.T) {
	w, r := getHandlerMockWithPath(t, "non-existing")
	handler, _, originalFallbackData := getWebserverHandler(t, []string{})
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, originalFallbackData, getReceivedData(t, w.Result().Body))
}

// TestWebServerSimpleServe check sif a plain file without any extras is delivered
func TestFileServerZip(t *testing.T) {
	w, r := getHandlerMockWithPath(t, testFile)
	r.Header.Set("Accept-Encoding", "gzip")
	handler, originalFileData, _ := getWebserverHandler(t, []string{"application/javascript"})
	originalFileDataZipped, err := utils.Zip(originalFileData, gzip.BestCompression)
	assert.Nil(t, err)
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, originalFileDataZipped, getReceivedData(t, w.Result().Body))
	assert.Equal(t, "gzip", w.Result().Header.Get("Content-Encoding"))
}

func TestFileServerZipFallback(t *testing.T) {
	w, r := getHandlerMockWithPath(t, "non-existing")
	r.Header.Set("Accept-Encoding", "gzip")
	handler, _, originalFallbackData := getWebserverHandler(t, []string{"text/html"})
	originalFileDataZipped, err := utils.Zip(originalFallbackData, gzip.BestCompression)
	assert.Nil(t, err)
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, originalFileDataZipped, getReceivedData(t, w.Result().Body))
	assert.Equal(t, "gzip", w.Result().Header.Get("Content-Encoding"))
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

func getHandlerMockWithPath(t *testing.T, path string) (responseWriter *httptest.ResponseRecorder, request *http.Request) {
	w, r, _ := getDefaultHandlerMocks()
	url, err := url.Parse(path)
	assert.Nil(t, err)
	r.URL = url
	return w, r
}
