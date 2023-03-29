package server

import (
	"github.com/felixge/httpsnoop"
	"github.com/ngergs/websrv/v3/internal/utils"
	"io"
	"net/http"
)

// FallbackHandler routes the request to a fallback route on of the given HTTP fallback status codes
func FallbackHandler(next http.Handler, fallbackPath string, fallbackCodes ...int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := 200
		wrappedW := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(headerFunc httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					status = code
					if !utils.Contains(fallbackCodes, code) {
						headerFunc(code)
					}
				}
			},
			Write: func(writeFunc httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) {
					if utils.Contains(fallbackCodes, status) {
						// dummy to avoid setting Content-Length here
						return len(b), nil
					}
					return writeFunc(b)
				}
			},
			ReadFrom: func(fromFunc httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(src io.Reader) (int64, error) {
					if utils.Contains(fallbackCodes, status) {
						// dummy to avoid setting Content-Length here
						b, err := io.ReadAll(src)
						return int64(len(b)), err
					}
					return fromFunc(src)
				}
			},
		})
		next.ServeHTTP(wrappedW, r)
		if utils.Contains(fallbackCodes, status) && r.URL.Path != fallbackPath {
			r.URL.Path = fallbackPath
			w.Header().Del("Content-Type")
			next.ServeHTTP(w, r)
		}
	})
}
