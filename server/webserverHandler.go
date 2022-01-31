package server

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
)

type WebserverHandler struct {
	fallbackFilepath string
	fileSystem       fs.FS
	config           *Config
	hashes           map[string]string
	templateServer   *templateServer
}

func New(fileSystem fs.FS, fallbackFilepath string, httpHeaderConfig *Config) (*WebserverHandler, error) {
	handler := &WebserverHandler{
		fallbackFilepath: fallbackFilepath,
		fileSystem:       fileSystem,
		config:           httpHeaderConfig,
		hashes:           make(map[string]string),
		templateServer: &templateServer{
			filesystems: fileSystem,
			templates:   make(map[string]*template.Template),
		},
	}
	// compute hashes
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		log.Debug().Msgf("Compute hash for %s", path)
		file, err := fileSystem.Open(path)
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
	if err != nil {
		return nil, fmt.Errorf("error initializing hashes: %w", err)
	}
	return handler, nil
}

func (handler *WebserverHandler) setHeaders(w http.ResponseWriter) {
	if handler.config != nil {
		for k, v := range handler.config.Headers {
			w.Header()[k] = v
		}
	}
	replacer := handler.config.RandomIdReplacer
	headerElements, ok := w.Header()[replacer.HeaderName]
	if !ok {
		return
	}
	for i, header := range headerElements {
		headerElements[i] = strings.Replace(header, replacer.VariableName, "todo", -1)
	}

}

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !handler.validate(w, r) {
		return
	}
	// remove leading / from path to make it relative
	// important to do this after cleaning, else relative paths may remain
	requestPath := path.Clean(r.URL.Path)[1:]

	ifNoneMatch := r.Header.Get("If-None-Match")
	/*if handler.checkIfNoneMatch(requestPath, ifNoneMatch) {
		handler.setHeaders(w)
		w.WriteHeader(http.StatusNotModified)
		return
	}*/

	if handler.config.RandomIdReplacer.FileNamePattern.MatchString(requestPath) {
		log.Debug().Msgf("Serving template file %s", requestPath)
		// needed due to https://github.com/golang/go/issues/32350
		w.Header()["Content-Type"] = []string{handler.getMediaType(requestPath)}
		handler.setHeaders(w)
		err := handler.templateServer.Serve(w, requestPath, map[string]string{handler.config.RandomIdReplacer.VariableName: "todo"})

		if err != nil {
			log.Warn().Err(err).Msg("error serving template file")
			http.Error(w, "failed to serve requested file, you can retry.", http.StatusInternalServerError)
		}
		return
	}

	file, err := handler.tryGetFile(requestPath)
	if err != nil {
		log.Debug().Err(err).Msgf("file %s not found", requestPath)
		var finishServing bool
		file, requestPath, finishServing = handler.checkForFallbackFile(w, requestPath, ifNoneMatch)
		if finishServing {
			return
		}
	}
	defer file.Close()

	// needed due to https://github.com/golang/go/issues/32350
	w.Header()["Content-Type"] = []string{handler.getMediaType(requestPath)}
	/*hash, ok := handler.hashes[requestPath]
	if ok {
		w.Header()["ETag"] = []string{hash}
	}*/
	// todo: make optional
	handler.setHeaders(w)
	if handler.getMediaType(requestPath) == "application/javascript" {
		_, err = io.Copy(w, file)
	} else {
		w.Header()["Content-Encoding"] = []string{"gzip"}
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		_, err = io.Copy(gzipWriter, file)
	}
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
}

func (handler *WebserverHandler) validate(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		http.Error(w, "This server only supports GET", http.StatusMethodNotAllowed)
		return false
	}
	if !path.IsAbs(r.URL.Path) {
		http.Error(w, "", http.StatusBadRequest)
		return false
	}
	return true
}

func (handler *WebserverHandler) tryGetFile(requestPath string) (fs.File, error) {
	file, err := handler.fileSystem.Open(requestPath)
	if err != nil {
		return nil, err
	}
	fileInfo, err := file.Stat()
	if fileInfo.IsDir() {
		defer file.Close()
		return nil, fmt.Errorf("requested file is directory")
	}
	return file, err
}

//checkIfNoneMatch returns true if a match occured
func (handler *WebserverHandler) checkForFallbackFile(w http.ResponseWriter, requestPath string, ifNoneMatch string) (file fs.File, requestpath string, finishServing bool) {
	// requested files do not fall back to index.html
	if handler.fallbackFilepath == "" || (path.Ext(requestPath) != "" && path.Ext(requestPath) != ".") {
		http.Error(w, "file not found", http.StatusNotFound)
		return nil, "", true
	}
	requestPath = handler.fallbackFilepath
	if handler.checkIfNoneMatch(requestPath, ifNoneMatch) {
		w.WriteHeader(http.StatusNotModified)
		return nil, "", true
	}
	file, err := handler.fileSystem.Open(handler.fallbackFilepath)
	if err != nil {
		log.Error().Err(err).Msg("fallback file not found")
		http.Error(w, "file not found", http.StatusNotFound)
		return nil, "", true
	}
	return file, requestPath, false
}

func (handler *WebserverHandler) checkIfNoneMatch(requestPath string, ifNoneMatch string) (match bool) {
	if ifNoneMatch != "" && ifNoneMatch == handler.hashes[requestPath] {
		hash, ok := handler.hashes[requestPath]
		if ok && ifNoneMatch == hash {
			return true
		}
	}
	return false
}

func (handler *WebserverHandler) getMediaType(requestPath string) string {
	mediaType, ok := handler.config.MediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	return mediaType
}
