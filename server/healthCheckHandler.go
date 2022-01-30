package server

import "net/http"

type HealthCheckHandler struct{}

func (*HealthCheckHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
