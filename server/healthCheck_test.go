package server_test

import (
	"github.com/ngergs/websrv/server"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckHandler()
	handler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestConditionalHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckConditionalHandler(func() bool { return true })
	handler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	w, r, _ = getDefaultHandlerMocks()
	handler = server.HealthCheckConditionalHandler(func() bool { return false })
	handler.ServeHTTP(w, r)
	require.Equal(t, http.StatusServiceUnavailable, w.Result().StatusCode)
}
