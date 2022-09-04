package server

import "net/http"

// HealthCheckHandler is a dummy handler that always returns HTTP 200.
func HealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// HealthCheckConditionalHandler is a conditional healthcheck handler that returns HTTP 200 when the condition argument function returns true and HTTP 503 if not.
func HealthCheckConditionalHandler(condition func() bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if condition() {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}
