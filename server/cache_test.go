package server_test

import (
	"github.com/ngergs/websrv/server"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEtagSetting(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.Nil(t, err)
	}
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	hash, ok := cacheHandler.Hashes.Get(path)
	assert.True(t, ok)
	assert.Equal(t, hash, w.Header().Get("ETag"))
}
func TestNoEtagOnError(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.Nil(t, err)
	}
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, "", w.Header().Get("ETag"))
}

func TestNotModifiedResponse(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r) // initial request to warm up the cache
	hash, ok := cacheHandler.Hashes.Get(path)
	assert.True(t, ok)
	r.Header.Set("If-None-Match", hash)
	w, _, _ = getDefaultHandlerMocks()
	cacheHandler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotModified, w.Result().StatusCode)
}
