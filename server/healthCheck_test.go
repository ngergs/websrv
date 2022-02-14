package server_test

import (
	"net/http"
	"testing"

	"github.com/ngergs/webserver/server"
)

func TestHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckHandler()
	w.mock.On("WriteHeader", http.StatusOK)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
}
