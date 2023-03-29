package server_test

import (
	"github.com/ngergs/websrv/v3/server"
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
	require.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}

func TestNonAbsolutePathRejection(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	url, err := url.Parse("../../../etc")
	require.Nil(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateHandler(next)
	handler.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestCleanPathTransversalAttack(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	// just a pattern to check that the path is correctly shortened
	url, err := url.Parse("/../../a/b/../c")
	require.Nil(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateHandler(next)
	handler.ServeHTTP(w, r)
	require.Equal(t, "/a/c", r.URL.Path)
}
