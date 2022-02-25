package server_test

import (
	"net/http"
	"testing"

	"github.com/ngergs/websrv/server"
	"github.com/stretchr/testify/assert"
)

const key = "abc"
const val = "test"

func TestHeaderHandler(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.HeaderHandler{Next: next, Headers: map[string]string{key: val}}
	var responseHeader http.Header = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assert.Equal(t, []string{val}, responseHeader.Values(key))
}
