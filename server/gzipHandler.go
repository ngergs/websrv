package server

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type gzipResponseWriter struct {
	Next           http.ResponseWriter
	GzipMediaTypes []string
	ZipWriter      *gzip.Writer
	selectedWriter io.Writer
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.Next.Header()
}

//containsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip")
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if w.selectedWriter == nil {
		if contains(w.GzipMediaTypes, w.Header().Get("Content-Type")) {
			w.selectedWriter = w.ZipWriter
			w.Header()["Content-Encoding"] = []string{"gzip"}
		} else {
			w.selectedWriter = w.Next
		}
	}
	return w.selectedWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Next.WriteHeader(statusCode)
}

func GzipHandler(next http.Handler, gzipMediaTypes []string) http.Handler {
	log.Debug().Msg("Adding gzip handler")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if containsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip") {
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()
			w = &gzipResponseWriter{Next: w, GzipMediaTypes: gzipMediaTypes, ZipWriter: gzipWriter}
		}
		next.ServeHTTP(w, r)
	})
}
