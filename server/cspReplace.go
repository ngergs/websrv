package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/ngergs/websrv/v3/internal/utils"
	"github.com/puzpuzpuz/xsync"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
)

var (
	ErrHttpStatusNotOk = errors.New("non HTTP 200 status code from underlying handler")
	ErrCachingTemplate = errors.New("error caching initial file for csp replacement")
)

// CspHeaderName is the Content-Security-Policy HTTP-Header name
const CspHeaderName = "Content-Security-Policy"

// CspFileHandler implements the http.Handler interface and fixes the Angular style-src CSP issue. The variableName is replaced
// in all response contents.
type CspFileHandler struct {
	replacer     *xsync.MapOf[string, *ReplacerCollection]
	Next         http.Handler
	VariableName string
	MediaTypeMap map[string]string
}

// NewCspFileHandler returns a CspFileHandler, it implements the http.Handler interface and fixes the Angular style-src CSP issue.
// The variableName is replaced in all response contents.
func NewCspFileHandler(next http.Handler, variableName string, mediaTypeMap map[string]string) *CspFileHandler {
	return &CspFileHandler{
		replacer:     xsync.NewMapOf[*ReplacerCollection](),
		Next:         next,
		VariableName: variableName,
		MediaTypeMap: mediaTypeMap,
	}
}

// CspHeaderHandler replaces the nonce placerholder in the Content-Security-header
func CspHeaderHandler(next http.Handler, variableName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := getSessionId(r)
		cspHeader := w.Header().Get(CspHeaderName)
		cspHeader = strings.ReplaceAll(cspHeader, variableName, sessionId)
		w.Header().Set(CspHeaderName, cspHeader)
		next.ServeHTTP(w, r)
	})
}

// loadTemplate loads a new template from the next handler
//
//nolint:contextcheck // context is obtained from request
func (handler *CspFileHandler) loadTemplate(w http.ResponseWriter, r *http.Request) (*ReplacerCollection, error) {
	status := http.StatusOK
	pr, pw := io.Pipe()
	defer utils.Close(r.Context(), pr)
	dummyHeader := make(http.Header)
	wrappedW := httpsnoop.Wrap(w, httpsnoop.Hooks{
		Header: func(_ httpsnoop.HeaderFunc) httpsnoop.HeaderFunc {
			return func() http.Header {
				// dummy to avoid setting Content-Length here
				return dummyHeader
			}
		},
		WriteHeader: func(headerFunc httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(code int) {
				status = code
				// forwarding would block setting headers later on
				if status != http.StatusOK {
					headerFunc(code)
				}
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
		return nil, fmt.Errorf("%w: %d", ErrHttpStatusNotOk, status)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCachingTemplate, err)
	}

	fileExtension := strings.Split(r.URL.Path, ".")
	mediaType, ok := handler.MediaTypeMap["."+fileExtension[len(fileExtension)-1]]
	if !ok {
		mediaType = "application/octet-stream"
	}
	replacer := ReplacerCollectionFromInput(data, handler.VariableName, mediaType)
	storedReplacer, _ := handler.replacer.LoadOrStore(r.URL.Path, replacer)
	return storedReplacer, nil
}

func (handler *CspFileHandler) serveFile(w http.ResponseWriter, r *http.Request, input string) error {
	replacer, ok := handler.replacer.Load(r.URL.Path)
	if !ok {
		var err error
		replacer, err = handler.loadTemplate(w, r)
		if err != nil {
			return err
		}
	}
	w.Header().Set("Content-Type", replacer.mediaType)
	return replacer.Replace(w, input)
}

func (handler *CspFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionId(r)
	err := handler.serveFile(w, r, sessionId)
	if err != nil {
		log.Err(err).Msgf("error serving template file %s", r.URL.Path)
		http.Error(w, "Error serving file.", http.StatusInternalServerError)
	}
}

// getSessionId extract the session id from the request context. Returns an empty string if it is not set.
func getSessionId(r *http.Request) string {
	sessionId := r.Context().Value(SessionIdKey)
	if sessionId == nil {
		log.Warn().Msg("SessionId not present in context")
		sessionId = "" // still replace to not leak the value that will be replaced
	}
	if sessionIdStr, ok := sessionId.(string); ok {
		return sessionIdStr
	} else {
		log.Warn().Msgf("sessionId is stored, but not a string: %v, will return empty string", sessionId)
		return ""
	}
}

// ReplacerCollection is a series of replacer implementations which are used to effectively replace a given string by pre-splitting the target template.
type ReplacerCollection struct {
	replacer  []replacer
	mediaType string
}

type replacer interface {
	Replace(w io.Writer, input string) error
}

type staticCopy struct {
	data []byte
}

type inputCopy struct{}

func (replacer *staticCopy) Replace(w io.Writer, _ string) error {
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
func ReplacerCollectionFromInput(data []byte, toReplace string, mediaType string) *ReplacerCollection {
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
	return &ReplacerCollection{replacer: replacer, mediaType: mediaType}
}
