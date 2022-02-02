package server

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog/log"
)

type CacheHandler struct {
	Next       http.Handler
	FileSystem fs.FS
	hashes     map[string]string
}

type eTagResponseWriter struct {
	Next       http.ResponseWriter
	Hash       string
	statusCode int
}

func (w *eTagResponseWriter) Header() http.Header {
	return w.Next.Header()
}

func (w *eTagResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 || w.statusCode == http.StatusOK {
		w.Header().Set("ETag", w.Hash)
	}
	return w.Next.Write(data)
}

func (w *eTagResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.Next.WriteHeader(statusCode)
}

func (handler *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "caching")

	eTag, ok := handler.hashes[r.URL.Path]
	if ok {
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch == eTag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		// Add ETag to 200 response
		w = &eTagResponseWriter{Next: w, Hash: eTag}
	}
	handler.Next.ServeHTTP(w, r)
}

// Init computes and stores the hashes for all files
func (handler *CacheHandler) Init() error {
	// compute hashes
	handler.hashes = make(map[string]string)
	return fs.WalkDir(handler.FileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		log.Debug().Msgf("Compute hash for %s", path)
		file, err := handler.FileSystem.Open(path)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(data)
		handler.hashes[path] = base64.StdEncoding.EncodeToString(hash[:])
		return nil
	})
}
