package server_test

import (
	"github.com/ngergs/websrv/v3/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
)

const dummyResponse = "hi"
const fallbackPath = "index.html"
const fallbackStatus = http.StatusNotFound
const fallbackResponse = "test123"

func TestNoFallback(t *testing.T) {
	// Setup test to get a session cookie
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fallbackPath {
			_, err := w.Write([]byte(fallbackResponse))
			require.NoError(t, err)
			return
		}
		_, err := w.Write([]byte(dummyResponse))
		require.NoError(t, err)
	}
	handler := server.FallbackHandler(next, fallbackPath, fallbackStatus)
	r.URL = &url.URL{Path: "/"}
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	response, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, dummyResponse, string(response))
}

func TestFallback(t *testing.T) {
	// Setup test to get a session cookie
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fallbackPath {
			_, err := w.Write([]byte(fallbackResponse))
			require.NoError(t, err)
			return
		}
		w.WriteHeader(fallbackStatus)
		_, err := w.Write([]byte(dummyResponse))
		require.NoError(t, err)
	}
	handler := server.FallbackHandler(next, fallbackPath, fallbackStatus)
	r.URL = &url.URL{Path: "/"}
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	response, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, fallbackResponse, string(response))
}
