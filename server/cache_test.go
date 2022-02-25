package server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/server"
	"github.com/stretchr/testify/assert"
)

func TestEtagSetting(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{}) // dummy write to trigger ETAg setting
	}
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	var responseHeader http.Header = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)
	cacheHandler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assert.Equal(t, cacheHandler.Hashes[path], responseHeader.Get("ETag"))
}
func TestNoEtagOnError(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte{}) // dummy write to trigger ETAg setting
	}
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	w.mock.On("WriteHeader", http.StatusAccepted).Return()
	cacheHandler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t) // call to "Header" has not been mocked -> ETag has not been set
}

func TestNotModifiedResponse(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	r.Header.Set("If-None-Match", cacheHandler.Hashes[path])
	w.mock.On("WriteHeader", http.StatusNotModified)
	cacheHandler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
}

func TestCacheMissCallsNext(t *testing.T) {
	path := "not_present.js"
	w, r, next := getDefaultHandlerMocks()
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, next.r, r)
	assert.Equal(t, next.w, w)
}

func getCacheHandler(t *testing.T, next http.Handler) *server.CacheHandler {
	fs, err := filesystem.NewMemoryFs("../benchmark")
	assert.Nil(t, err)
	cacheHandler, err := server.NewCacheHandler(next, fs)
	assert.Nil(t, err)
	return cacheHandler
}
