package server

import (
	"net/http"
)

type HeaderHandler struct {
	Next    http.Handler
	Headers map[string]string
}

func (handler *HeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "header")

	// set static headers
	if handler.Headers != nil {
		for k, v := range handler.Headers {
			w.Header().Set(k, v)
		}
	}
	handler.Next.ServeHTTP(w, r)
}
