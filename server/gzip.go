package server

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/ngergs/websrv/internal/utils"
)

type gzipResponseWriter struct {
	Next             http.ResponseWriter
	selectedWriter   io.Writer
	zipWriter        *gzip.Writer
	GzipMediaTypes   []string
	CompressionLevel int
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.Next.Header()
}

func (w *gzipResponseWriter) initWriter() error {
	if w.Header().Get("Content-Encoding") == "" &&
		utils.Contains(w.GzipMediaTypes, w.Header().Get("Content-Type")) {
		zipWriter, err := gzip.NewWriterLevel(w.Next, w.CompressionLevel)
		if err != nil {
			return err
		}
		w.zipWriter = zipWriter
		w.selectedWriter = w.zipWriter
		w.Header().Set("Content-Encoding", "gzip")
	} else {
		w.selectedWriter = w.Next
	}
	return nil
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if w.selectedWriter == nil {
		err := w.initWriter()
		if err != nil {
			return 0, err
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

// GzipHandler is a handler that compressed responses if the HTTP Content-Type response header is part of the gzipMediaTypes
// and if the request hat the Accept-Encoding=gzip HTTP request Header set.
func GzipHandler(next http.Handler, compressionLevel int, gzipMediaTypes []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "gzip")
		if utils.ContainsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip") {
			gzipResponseWriter := &gzipResponseWriter{
				Next:             w,
				CompressionLevel: compressionLevel,
				GzipMediaTypes:   gzipMediaTypes}
			defer utils.Close(r.Context(), gzipResponseWriter)
			w = gzipResponseWriter
		}
		next.ServeHTTP(w, r)
	})
}
