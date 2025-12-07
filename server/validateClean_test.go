package server_test

import (
	"github.com/ngergs/websrv/v4/server"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWrongMethod(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	r.Method = http.MethodPost
	handler := server.ValidateHandler(next)
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusMethodNotAllowed, result.StatusCode)
}

func TestNonAbsolutePathRejection(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	url, err := url.Parse("../../../etc")
	require.NoError(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateHandler(next)
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusBadRequest, result.StatusCode)
}

func TestCleanPathTransversalAttack(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	// just a pattern to check that the path is correctly shortened
	url, err := url.Parse("/../../a/b/../c")
	require.NoError(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateHandler(next)
	handler.ServeHTTP(w, r)
	defer func() {
		err := w.Result().Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, "/a/c", r.URL.Path)
}
