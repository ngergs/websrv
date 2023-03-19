package server

import (
	"net/http"
)

// HeaderHandler implements the http.Handler interface and adds the static headers provided in the Headers map to the response.
type HeaderHandler struct {
	Next    http.Handler
	Headers map[string]string
}

func (handler *HeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set static headers
	if handler.Headers != nil {
		for k, v := range handler.Headers {
			w.Header().Set(k, v)
		}
	}
	handler.Next.ServeHTTP(w, r)
}
