package server

import (
	"net/http"
	"path"
)

// ValidateHandler returns HTTP 405 if the request method is not GET or HEAD.
// Also, pa relative paths are rejected with HTTP 400.
func ValidateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "This server only supports HTTP methods GET and HEAD", http.StatusMethodNotAllowed)
			return
		}
		if !path.IsAbs(r.URL.Path) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		r.URL.Path = path.Clean(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
