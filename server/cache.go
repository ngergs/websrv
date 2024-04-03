package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"github.com/felixge/httpsnoop"
	"github.com/ngergs/websrv/v3/internal/syncwrap"
	"github.com/ngergs/websrv/v3/internal/utils"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

// cacheHandler implements a http.Handler that supports Caching via the ETag and If-None-Match HTTP-Headers.
// The CacheHandler required that all following handlers only serve static resources.
// The next handler in the chain is only called when a cache mismatch occurs.
type cacheHandler struct {
	Next   http.Handler
	Hashes *syncwrap.Map[string, string]
}

//nolint:contextcheck // context is obtained from request
func (handler *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	status := http.StatusOK
	pr, pw := io.Pipe()
	defer utils.Close(r.Context(), pr)
	wrappedW := httpsnoop.Wrap(w, httpsnoop.Hooks{
		WriteHeader: func(headerFunc httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(code int) {
				status = code
				headerFunc(code)
			}
		},
		Write: func(writeFunc httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(b []byte) (int, error) {
				if status == http.StatusOK {
					return pw.Write(b)
				}
				return writeFunc(b)
			}
		},
		ReadFrom: func(fromFunc httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
			return func(src io.Reader) (int64, error) {
				if status == http.StatusOK {
					return io.Copy(pw, src)
				}
				return fromFunc(src)
			}
		},
	})
	go func() {
		handler.Next.ServeHTTP(wrappedW, r)
		utils.Close(r.Context(), pw)
	}()
	data, err := io.ReadAll(pr)
	if status != http.StatusOK {
		return
	}
	if err != nil {
		log.Err(err).Msgf("error storing response in middleware to determine hash %s", r.URL.Path)
		http.Error(w, "Error serving file.", http.StatusInternalServerError)
	}
	hash := sha256.Sum256(data)
	eTag = base64.StdEncoding.EncodeToString(hash[:])
	log.Debug().Msgf("Computed missing eTag for %s: %s", r.URL.Path, eTag)
	handler.Hashes.Set(r.URL.Path, eTag)
	w.Header().Set("ETag", eTag)

	_, err = io.Copy(w, bytes.NewReader(data))
	if err != nil {
		log.Err(err).Msgf("error coping response in middleware after determining hash %s", r.URL.Path)
	}
}

// NewCacheHandler computes and stores the hashes for all files
func NewCacheHandler(next http.Handler) *cacheHandler {
	// compute hashes
	return &cacheHandler{
		Next:   next,
		Hashes: syncwrap.NewMap[string, string](),
	}
}
