package server_test

import (
	"github.com/ngergs/websrv/server"
	"net/http"
	"net/url"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestEtagSetting(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.Nil(t, err)
	}
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, cacheHandler.Hashes[path], w.Result().Header.Get("ETag"))
}
func TestNoEtagOnError(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.Nil(t, err)
	}
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, "", w.Result().Header.Get("ETag"))
}

func TestNotModifiedResponse(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	cacheHandler := getCacheHandler(t, next)
	r.URL = &url.URL{Path: path}
	r.Header.Set("If-None-Match", cacheHandler.Hashes[path])
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotModified, w.Result().StatusCode)
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
	fs, err := filesystem.NewMemoryFs("../test/benchmark")
	assert.Nil(t, err)
	cacheHandler, err := server.NewCacheHandler(next, fs)
	assert.Nil(t, err)
	return cacheHandler
}
