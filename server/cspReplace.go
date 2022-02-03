package server

import (
	"net/http"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/ngergs/webserver/v2/filesystem"
	"github.com/ngergs/webserver/v2/utils"
	"github.com/rs/zerolog/log"
)

const cspHeaderName = "Content-Security-Policy"

type CspReplaceHandler struct {
	Next           http.Handler
	Filesystem     filesystem.ZipFs
	FileNamePatter *regexp.Regexp
	VariableName   string
	Templates      map[string]*template.Template
	MediaTypeMap   map[string]string
}

func (handler *CspReplaceHandler) load(path string) (*template.Template, error) {
	data, err := handler.Filesystem.ReadFile(path)
	if err != nil {
		return nil, err
	}
	isZipped, err := handler.Filesystem.IsZipped(path)
	if err != nil {
		return nil, err
	}
	if isZipped {
		data, err = utils.Unzip(data)
		if err != nil {
			return nil, err
		}
	}
	template, err := template.New(path).Delims("[{[{", "}]}]").Parse(string(data))
	if err != nil {
		return nil, err
	}
	handler.Templates[path] = template
	return template, nil
}

func (handler *CspReplaceHandler) serveFile(w http.ResponseWriter, path string, data interface{}) error {
	template, ok := handler.Templates[path]
	if !ok {
		var err error
		template, err = handler.load(path)
		if err != nil {
			return err
		}
	}
	template, err := template.Clone()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", handler.getMediaType(path))
	return template.Execute(w, data)
}

func (handler *CspReplaceHandler) replaceHeader(w http.ResponseWriter, sessionId string) {
	cspHeader := w.Header().Get(cspHeaderName)
	cspHeader = strings.Replace(cspHeader, handler.VariableName, sessionId, -1)
	w.Header().Set(cspHeaderName, cspHeader)
}

func (handler *CspReplaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "csp-replace")

	sessionId := r.Context().Value(sessionIddKey)
	if sessionId == nil {
		log.Warn().Msg("SessionId not present in context")
		sessionId = "" // stil replace to not leak the value that will be replaced
	}

	handler.replaceHeader(w, sessionId.(string))
	if !handler.FileNamePatter.MatchString(r.URL.Path) {
		handler.Next.ServeHTTP(w, r)
		return
	}
	err := handler.serveFile(w, r.URL.Path, map[string]string{handler.VariableName: sessionId.(string)})
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
