package server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ngergs/websrv/v3/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEtagSetting(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.NoError(t, err)
	}
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, result.StatusCode)
	hash, ok := cacheHandler.Hashes.Load(path)
	require.True(t, ok)
	require.Equal(t, hash, w.Header().Get("ETag"))
}
func TestNoEtagOnError(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte{}) // dummy write to trigger ETAg setting
		assert.NoError(t, err)
	}
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r)
	require.Empty(t, w.Header().Get("ETag"))
}

func TestNotModifiedResponse(t *testing.T) {
	path := "dummy_random.js"
	w, r, next := getDefaultHandlerMocks()
	cacheHandler := server.NewCacheHandler(next)
	r.URL = &url.URL{Path: path}
	cacheHandler.ServeHTTP(w, r) // initial request to warm up the cache
	hash, ok := cacheHandler.Hashes.Load(path)
	require.True(t, ok)
	r.Header.Set("If-None-Match", hash)
	w, _, _ = getDefaultHandlerMocks()
	cacheHandler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusNotModified, result.StatusCode)
}
