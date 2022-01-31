package server

import (
	"io"
	"io/fs"
	"text/template"
)

type templateServer struct {
	templates   map[string]*template.Template
	filesystems fs.FS
}

func (server *templateServer) load(path string) (*template.Template, error) {
	template := template.New(path)
	template, err := template.New(path).Delims("[{[{", "}]}]").ParseFS(server.filesystems, path)
	if err != nil {
		return nil, err
	}
	server.templates[path] = template
	return template, nil
}

func (server *templateServer) Serve(w io.Writer, path string, data interface{}) error {
	template, ok := server.templates[path]
	if !ok {
		var err error
		template, err = server.load(path)
		if err != nil {
			return err
		}
	}
	return template.Execute(w, data)
}
