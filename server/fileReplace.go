package server

import (
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"text/template"

	"github.com/rs/zerolog/log"
)

type FileReplaceHandler struct {
	Next             http.Handler
	Filesystem       fs.FS
	SourceHeaderName string
	FileNamePatter   *regexp.Regexp
	VariableName     string
	Templates        map[string]*template.Template
	MediaTypeMap     map[string]string
}

func (handler *FileReplaceHandler) load(path string) (*template.Template, error) {
	template := template.New(path)
	template, err := template.New(path).Delims("[{[{", "}]}]").ParseFS(handler.Filesystem, path)
	if err != nil {
		return nil, err
	}
	handler.Templates[path] = template
	return template, nil
}

func (handler *FileReplaceHandler) Serve(w http.ResponseWriter, path string, data interface{}) error {
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

func (handler *FileReplaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Ctx(r.Context()).Debug().Msg("Entering file replace handler")
	if !handler.FileNamePatter.MatchString(r.URL.Path) {
		handler.Next.ServeHTTP(w, r)
		return
	}
	err := handler.Serve(w, r.URL.Path, map[string]string{handler.VariableName: getLastHeaderEntry(r, handler.SourceHeaderName)})
	if err != nil {
		log.Err(err).Msgf("error serving template file %s", r.URL.Path)
		http.Error(w, "Error serving file.", http.StatusInternalServerError)
	}

}

func (handler *FileReplaceHandler) getMediaType(requestPath string) string {
	mediaType, ok := handler.MediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	return mediaType
}
