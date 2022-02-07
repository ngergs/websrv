package server

import (
	"context"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

const cspHeaderName = "Content-Security-Policy"

type CspReplaceHandler struct {
	Next           http.Handler
	Filesystem     fs.ReadFileFS
	FileNamePatter *regexp.Regexp
	VariableName   string
	Replacer       map[string]*replacerCollection
	MediaTypeMap   map[string]string
	ReplacerLock   sync.RWMutex
}

func (handler *CspReplaceHandler) load(path string) (*replacerCollection, error) {
	data, err := handler.Filesystem.ReadFile(path)
	if err != nil {
		return nil, err
	}
	replacer, err := ReplacerCollectionFromInput(data, handler.VariableName)
	if err != nil {
		return nil, err
	}
	handler.ReplacerLock.Lock()
	defer handler.ReplacerLock.Unlock()
	handler.Replacer[path] = replacer
	return replacer, nil
}

func (handler *CspReplaceHandler) getTemplate(path string) (replacer *replacerCollection, ok bool) {
	handler.ReplacerLock.RLock()
	defer handler.ReplacerLock.RUnlock()
	replacer, ok = handler.Replacer[path]
	return
}

func (handler *CspReplaceHandler) serveFile(ctx context.Context, w http.ResponseWriter, path string, input string) error {
	replacer, ok := handler.getTemplate(path)
	if !ok {
		var err error
		replacer, err = handler.load(path)
		if err != nil {
			return err
		}
	}
	w.Header().Set("Content-Type", handler.getMediaType(path))
	return replacer.Replace(ctx, w, input)
}

func (handler *CspReplaceHandler) replaceHeader(w http.ResponseWriter, sessionId string) {
	cspHeader := w.Header().Get(cspHeaderName)
	cspHeader = strings.Replace(cspHeader, handler.VariableName, sessionId, -1)
	w.Header().Set(cspHeaderName, cspHeader)
}

func (handler *CspReplaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "csp-replace")

	sessionId := r.Context().Value(SessionIdKey)
	if sessionId == nil {
		log.Warn().Msg("SessionId not present in context")
		sessionId = "" // stil replace to not leak the value that will be replaced
	}

	handler.replaceHeader(w, sessionId.(string))
	if handler.FileNamePatter == nil {
		log.Ctx(r.Context()).Warn().Msg("Csp-Replace handler invoced but FileNamePattern has not been set. Skipping file replacement")
		handler.Next.ServeHTTP(w, r)
		return
	}
	if !handler.FileNamePatter.MatchString(r.URL.Path) {
		handler.Next.ServeHTTP(w, r)
		return
	}
	err := handler.serveFile(r.Context(), w, r.URL.Path, sessionId.(string))
	if err != nil {
		log.Err(err).Msgf("error serving template file %s", r.URL.Path)
		http.Error(w, "Error serving file.", http.StatusInternalServerError)
	}
}

func (handler *CspReplaceHandler) getMediaType(requestPath string) string {
	mediaType, ok := handler.MediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	return mediaType
}
