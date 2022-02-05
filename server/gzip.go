package server

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/ngergs/webserver/v2/utils"
)

type gzipResponseWriter struct {
	Next           http.ResponseWriter
	GzipMediaTypes []string
	zipWriter      *gzip.Writer
	selectedWriter io.Writer
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.Next.Header()
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if w.selectedWriter == nil {
		if w.Header().Get("Content-Encoding") == "" &&
			utils.Contains(w.GzipMediaTypes, w.Header().Get("Content-Type")) {
			w.zipWriter = gzip.NewWriter(w.Next)
			w.selectedWriter = w.zipWriter
			w.Header().Set("Content-Encoding", "gzip")
		} else {
			w.selectedWriter = w.Next
		}
	}
	return w.selectedWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Next.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Close() error {
	if w.zipWriter == nil {
		return nil
	}
	return w.zipWriter.Close()
}

func GzipHandler(next http.Handler, gzipMediaTypes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "gzip")
		if utils.ContainsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip") {
			gzipResponseWriter := &gzipResponseWriter{Next: w, GzipMediaTypes: gzipMediaTypes}
			defer utils.Close(r.Context(), gzipResponseWriter)
			w = gzipResponseWriter
		}
		next.ServeHTTP(w, r)
	})
}
