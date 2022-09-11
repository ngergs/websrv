package server_test

import (
	"github.com/ngergs/websrv/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

const key = "abc"
const val = "test"

func TestHeaderHandler(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.HeaderHandler{Next: next, Headers: map[string]string{key: val}}
	handler.ServeHTTP(w, r)
	assert.Equal(t, val, w.Result().Header.Get(key))
}
