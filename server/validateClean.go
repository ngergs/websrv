package server

import (
	"net/http"
	"path"

	"github.com/rs/zerolog/log"
)

func ValidateCleanHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Ctx(r.Context()).Debug().Msg("Entering validate and clean handler")
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
