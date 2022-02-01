package server

import (
	"io"
	"io/fs"
	"net/http"
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
	templates        map[string]*template.Template
}

func (handler *FileReplaceHandler) load(path string) (*template.Template, error) {
	template := template.New(path)
	template, err := template.New(path).Delims("[{[{", "}]}]").ParseFS(handler.Filesystem, path)
	if err != nil {
		return nil, err
	}
	handler.templates[path] = template
	return template, nil
}

func (handler *FileReplaceHandler) Serve(w io.Writer, path string, data interface{}) error {
	template, ok := handler.templates[path]
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
	return template.Execute(w, data)
}

func (handler *FileReplaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
