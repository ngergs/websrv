package server

import (
	"net/http"
	"path"
)

// ValidateCleanHandler returns HTTP 405 if the request method is not GET or HEAD.
// Also relative paths are rejected with HTTP 400.
func ValidateCleanHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "validate-clean")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "This server only supports HTTP methods GET and HEAD", http.StatusMethodNotAllowed)
			return
		}
		if !path.IsAbs(r.URL.Path) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// remove leading / from path to make it relative
		// important to do this after cleaning, else relative paths may remain
		r.URL.Path = path.Clean(r.URL.Path)[1:]
		next.ServeHTTP(w, r)
	})
}
