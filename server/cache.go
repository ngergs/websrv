package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"github.com/ngergs/websrv/internal/sync"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

// cacheHandler implements a http.Handler that supports Caching via the ETag and If-None-Match HTTP-Headers.
// The CacheHandler required that all following handlers only serve static resources.
// The next handler in the chain is only called when a cache mismatch occurs.
type cacheHandler struct {
	Next   http.Handler
	Hashes *sync.Map[string, string]
}

// bufferedResponseWriter buffers the output response (to calculate the hash) but passes status codes and headers just through
type bufferedResponseWriter struct {
	Next       http.ResponseWriter
	StatusCode int
	Buffer     bytes.Buffer
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.Next.Header()
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.Next.WriteHeader(statusCode)
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	return w.Next.Write(data)
}

func (handler *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "caching")

	eTag, ok := handler.Hashes.Get(r.URL.Path)
	if ok {
		if r.Header.Get("If-None-Match") == eTag {
			log.Debug().Msgf("Returned not modified for %s: %s", r.URL.Path, eTag)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		// we have the hash but not present in the request, add e-tag and continue
		log.Debug().Msgf("Returned already stored eTag for %s: %s", r.URL.Path, eTag)
		w.Header().Set("ETag", eTag)
		handler.Next.ServeHTTP(w, r)
		return
	}

	// We do not have the hash yet, get it and add ETag
	bufferedW := &bufferedResponseWriter{Next: w}
	handler.Next.ServeHTTP(bufferedW, r)
	if bufferedW.StatusCode == 0 || bufferedW.StatusCode == http.StatusOK {
		hash := sha256.Sum256(bufferedW.Buffer.Bytes())
		eTag = base64.StdEncoding.EncodeToString(hash[:])
		log.Debug().Msgf("Computed missing eTag for %s: %s", r.URL.Path, eTag)
		handler.Hashes.Set(r.URL.Path, eTag)
		w.Header().Set("ETag", eTag)
	}
	io.Copy(w, &bufferedW.Buffer)
}

// NewCacheHandler computes and stores the hashes for all files
func NewCacheHandler(next http.Handler) *cacheHandler {
	// compute hashes
	return &cacheHandler{
		Next:   next,
		Hashes: sync.NewMap[string, string](),
	}
}
