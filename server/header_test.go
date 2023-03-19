package server_test

import (
	"github.com/ngergs/websrv/v2/server"
	"testing"

	"github.com/stretchr/testify/require"
)

const key = "abc"
const val = "test"

func TestHeaderHandler(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.HeaderHandler{Next: next, Headers: map[string]string{key: val}}
	handler.ServeHTTP(w, r)
	require.Equal(t, val, w.Result().Header.Get(key))
}
