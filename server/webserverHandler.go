package server

import (
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

func New(fileSystem fs.FS, fallbackFilepath string, config *Config) (*WebserverHandler, error) {
	handler := &WebserverHandler{
		fallbackFilepath: fallbackFilepath,
		fileSystem:       fileSystem,
		config:           config,
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

func (handler *WebserverHandler) setHeaders(w http.ResponseWriter, replacement string) {
	if handler.config != nil {
		for k, v := range handler.config.Headers {
			w.Header().Set(k, v)
		}
	}

	// nonce replacement if necessary
	replacer := handler.config.FromHeaderReplacer
	if replacer == nil {
		return
	}
	header := w.Header().Get(replacer.TargetHeaderName)
	if header == "" {
		return
	}
	w.Header().Set(replacer.TargetHeaderName, strings.Replace(header, replacer.VariableName, replacement, -1))
}

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !handler.validate(w, r) {
		return
	}
	var fromHeaderReplacement string
	if handler.config.FromHeaderReplacer != nil {
		headers := r.Header.Values(handler.config.FromHeaderReplacer.SourceHeaderName)
		if len(headers) > 0 {
			fromHeaderReplacement = headers[len(headers)-1]
		}
	}

	// remove leading / from path to make it relative
	// important to do this after cleaning, else relative paths may remain
	requestPath := path.Clean(r.URL.Path)[1:]

	var serve func(w http.ResponseWriter) error
	if handler.config.FromHeaderReplacer != nil &&
		handler.config.FromHeaderReplacer.FileNamePattern.MatchString(requestPath) {
		log.Debug().Msgf("Serving template file %s", requestPath)
		// needed due to https://github.com/golang/go/issues/32350
		serve = func(w http.ResponseWriter) error {
			return handler.templateServer.Serve(w, requestPath, map[string]string{handler.config.FromHeaderReplacer.VariableName: fromHeaderReplacement})
		}
	} else {
		ifNoneMatch := r.Header.Get("If-None-Match")
		if handler.chechHash(requestPath, ifNoneMatch) {
			handler.setHeaders(w, fromHeaderReplacement)
			w.WriteHeader(http.StatusNotModified)
			return
		}

		log.Debug().Msgf("Serving file %s", requestPath)
		file, err := handler.tryGetFile(requestPath)
		if err == nil {
			// hash only for regular served files, not for fallback index file as this ruins CSP header
			hash, ok := handler.hashes[requestPath]
			if ok {
				w.Header()["ETag"] = []string{hash}
			}
		}
		if err != nil {
			log.Debug().Err(err).Msgf("file %s not found", requestPath)
			var finishServing bool
			file, requestPath, finishServing = handler.checkForFallbackFile(w, requestPath, ifNoneMatch)
			if finishServing {
				return
			}
		}
		defer file.Close()
		serve = func(w http.ResponseWriter) error {
			_, err := io.Copy(w, file)
			return err
		}
	}

	mediaType := handler.getMediaType(requestPath)
	w.Header()["Content-Type"] = []string{mediaType}
	handler.setHeaders(w, fromHeaderReplacement)

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	err := serve(w)
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
}

func (handler *WebserverHandler) validate(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "This server only supports HTTP methods GET and HEAD", http.StatusMethodNotAllowed)
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
	if handler.chechHash(requestPath, ifNoneMatch) {
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

func (handler *WebserverHandler) chechHash(requestPath string, ifNoneMatch string) (match bool) {
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
