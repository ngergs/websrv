package server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ngergs/webserver/server"
	"github.com/stretchr/testify/assert"
)

func TestValidateWrongMethod(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	w.mockStatusWrite(405)
	r.Method = http.MethodPost
	handler := server.ValidateCleanHandler(next)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
}

func TestNonAbsolutePathRejection(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	url, err := url.Parse("../../../etc")
	assert.Nil(t, err)
	r.Method = http.MethodGet
	r.URL = url
	w.mockStatusWrite(400)
	handler := server.ValidateCleanHandler(next)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
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
