package server_test

import (
	"github.com/ngergs/websrv/server"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateWrongMethod(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	r.Method = http.MethodPost
	handler := server.ValidateCleanHandler(next)
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}

func TestNonAbsolutePathRejection(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	url, err := url.Parse("../../../etc")
	assert.Nil(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateCleanHandler(next)
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestCleanPathTransversalAttack(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	// just a pattern to check that the path is correctly shortened
	url, err := url.Parse("/../../a/b/../c")
	assert.Nil(t, err)
	r.Method = http.MethodGet
	r.URL = url
	handler := server.ValidateCleanHandler(next)
	handler.ServeHTTP(w, r)
	assert.Equal(t, "a/c", next.r.URL.Path)
}
