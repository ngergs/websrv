package server_test

import (
	"github.com/ngergs/websrv/v2/server"
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
			w.Write([]byte(fallbackResponse))
			return
		}
		w.Write([]byte(dummyResponse))
	}
	handler := server.FallbackHandler(next, fallbackPath, fallbackStatus)
	r.URL = &url.URL{Path: "/"}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	response, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, dummyResponse, string(response))
}

func TestFallback(t *testing.T) {
	// Setup test to get a session cookie
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fallbackPath {
			w.Write([]byte(fallbackResponse))
			return
		}
		w.WriteHeader(fallbackStatus)
		w.Write([]byte(dummyResponse))
	}
	handler := server.FallbackHandler(next, fallbackPath, fallbackStatus)
	r.URL = &url.URL{Path: "/"}
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	response, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, fallbackResponse, string(response))
}