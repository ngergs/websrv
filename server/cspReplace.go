package server

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// CspHeaderName is the Content-Security-Policy HTTP-Header name
const CspHeaderName = "Content-Security-Policy"

// CspReplaceHandler implements the http.Handler interface and fixes the Angular style-src CSP issue. The variableName is replaced
// in files that match the FileNamePattern as well as in the Content-Security-Policy header.
type CspReplaceHandler struct {
	Next           http.Handler
	Filesystem     fs.ReadFileFS
	FileNamePatter *regexp.Regexp
	VariableName   string
	Replacer       map[string]*ReplacerCollection
	MediaTypeMap   map[string]string
	ReplacerLock   sync.RWMutex
}

func (handler *CspReplaceHandler) load(path string) (*ReplacerCollection, error) {
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

func (handler *CspReplaceHandler) getTemplate(path string) (replacer *ReplacerCollection, ok bool) {
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
	return replacer.Replace(w, input)
}

func (handler *CspReplaceHandler) replaceHeader(w http.ResponseWriter, sessionId string) {
	cspHeader := w.Header().Get(CspHeaderName)
	cspHeader = strings.Replace(cspHeader, handler.VariableName, sessionId, -1)
	w.Header().Set(CspHeaderName, cspHeader)
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

// ReplacerCollection is a series of replacer implementations which are used to effectively replace a given string by pre-splitting the target template.
type ReplacerCollection struct {
	replacer []replacer
}

type replacer interface {
	Replace(w io.Writer, input string) error
}

type staticCopy struct {
	data []byte
}

type inputCopy struct{}

func (replacer *staticCopy) Replace(w io.Writer, input string) error {
	r := bytes.NewReader(replacer.data)
	_, err := io.Copy(w, r)
	return err
}

func (replacer *inputCopy) Replace(w io.Writer, input string) error {
	data := []byte(input)
	r := bytes.NewReader(data)
	_, err := io.Copy(w, r)
	return err
}

// Replace replaces the template placeholder with the input string and writes the result to the io.Writer w.
func (replacer *ReplacerCollection) Replace(w io.Writer, input string) error {
	for _, subreplacer := range replacer.replacer {
		err := subreplacer.Replace(w, input)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReplacerCollectionFromInput constructs a replacer that prepares the input data into a template where the toReplace string will be replaced.
func ReplacerCollectionFromInput(data []byte, toReplace string) (*ReplacerCollection, error) {
	fragments := strings.Split(string(data), toReplace)
	// build replacement chain
	replacer := make([]replacer, 0)
	for i := 0; i < len(fragments)-1; i++ {
		data = []byte(fragments[i])
		replacer = append(replacer, &staticCopy{data: data})
		replacer = append(replacer, &inputCopy{})
	}
	data = []byte(fragments[len(fragments)-1])
	replacer = append(replacer, &staticCopy{data: data})
	return &ReplacerCollection{replacer: replacer}, nil
}
