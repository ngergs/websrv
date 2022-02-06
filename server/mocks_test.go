package server_test

import (
	"context"
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

type mockHandler struct {
	w http.ResponseWriter
	r *http.Request
}

func (handler *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.w = w
	handler.r = r
}

// getDefaultHandlerMocks provides default mocks used for handler testing
func getDefaultHandlerMocks() (w *mockResponseWriter, r *http.Request, next *mockHandler) {
	next = &mockHandler{}
	w = &mockResponseWriter{}
	r = &http.Request{Header: make(map[string][]string)}
	r = r.WithContext(context.Background())
	var responseHeader http.Header = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)
	return
}
