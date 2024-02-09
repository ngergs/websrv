package server_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHandler struct {
	w             http.ResponseWriter
	r             *http.Request
	serveHttpFunc func(w http.ResponseWriter, r *http.Request)
}

func (handler *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.w = w
	handler.r = r
	if handler.serveHttpFunc != nil {
		handler.serveHttpFunc(w, r)
	}
}

// getDefaultHandlerMocks provides default mocks used for handler testing
func getDefaultHandlerMocks() (w *httptest.ResponseRecorder, r *http.Request, next *mockHandler) {
	next = &mockHandler{}
	w = httptest.NewRecorder()
	r = &http.Request{Header: make(map[string][]string)}
	r = r.WithContext(context.Background())
	return
}

func getReceivedData(t *testing.T, r io.Reader) []byte {
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return data
}
