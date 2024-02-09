package server_test

import (
	"github.com/ngergs/websrv/v3/server"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckHandler()
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, result.StatusCode)
}

func TestConditionalHealthCheck(t *testing.T) {
	w, r, _ := getDefaultHandlerMocks()
	handler := server.HealthCheckConditionalHandler(func() bool { return true })
	handler.ServeHTTP(w, r)
	result := w.Result()
	defer func() {
		err := result.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, result.StatusCode)

	w, r, _ = getDefaultHandlerMocks()
	handler = server.HealthCheckConditionalHandler(func() bool { return false })
	handler.ServeHTTP(w, r)
	result2 := w.Result()
	defer func() {
		err := result2.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusServiceUnavailable, result2.StatusCode)
}
