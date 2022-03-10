package server_test

import (
	"net/http"
	"testing"

	"github.com/ngergs/websrv/server"
)

func TestHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckHandler()
	w.mock.On("WriteHeader", http.StatusOK)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
}

func TestConditionalHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckConditionalHandler(func() bool { return true })
	w.mock.On("WriteHeader", http.StatusOK)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)

	w, r, _ = getDefaultHandlerMocks()
	handler = server.HealthCheckConditionalHandler(func() bool { return false })
	w.mock.On("WriteHeader", http.StatusServiceUnavailable)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
}
