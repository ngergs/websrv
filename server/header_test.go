package server_test

import (
	"github.com/ngergs/websrv/v3/server"
	"testing"

	"github.com/stretchr/testify/require"
)

const key = "abc"
const val = "test"

func TestHeaderHandler(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.HeaderHandler{Next: next, Headers: map[string]string{key: val}}
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, val, result.Header.Get(key))
}
