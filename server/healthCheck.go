package server

import "net/http"

// HealthCheckHandler is a dummy handler that always returns HTTP 200.
func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
