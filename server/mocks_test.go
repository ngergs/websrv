package server_test

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockResponseWriter struct {
	mock mock.Mock
}

func (w *mockResponseWriter) Header() http.Header {
	args := w.mock.Called()
	return args.Get(0).(http.Header)
}

func (w *mockResponseWriter) Write(data []byte) (int, error) {
	args := w.mock.Called(data)
	return args.Int(0), args.Error(1)
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.mock.Called(statusCode)
}
