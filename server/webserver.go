package server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"text/template"

	"github.com/ngergs/webserver/v2/filesystem"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type WebserverHandler struct {
	fallbackFilepath string
	fileSystem       filesystem.ZipFs
	config           *Config
	templateServer   *FileReplaceHandler
	gzipMediaTypes   []string
}

func FileServerHandler(fileSystem filesystem.ZipFs, fallbackFilepath string, config *Config, hasMemoryFs bool) *WebserverHandler {
	handler := &WebserverHandler{
		fallbackFilepath: fallbackFilepath,
		fileSystem:       fileSystem,
		config:           config,
		templateServer: &FileReplaceHandler{
			Filesystem: fileSystem,
			Templates:  make(map[string]*template.Template),
		},
	}
	if hasMemoryFs {
		handler.gzipMediaTypes = config.GzipMediaTypes
	}
	return handler
}

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "webserver")
	logger := log.Ctx(r.Context())
	logger.Debug().Msg("Entering webserver handler")
	requestPath := r.URL.Path
	logger.Debug().Msgf("Serving file %s", requestPath)

	file, err := handler.tryGetFile(requestPath)
	if err != nil {
		logger.Debug().Err(err).Msgf("file %s not found", requestPath)
		var finishServing bool
		file, requestPath, finishServing = handler.checkForFallbackFile(logger, w, requestPath)
		if finishServing {
			return
		}
	}
	defer file.Close()

	mediaType := handler.getMediaType(requestPath)
	w.Header().Set("Content-Type", mediaType)
	isZipped, err := handler.fileSystem.IsZipped(requestPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to resolve whether file is zipped")
		http.Error(w, "file not found", http.StatusNotFound)
	}
	if isZipped {
		w.Header().Set("Content-Encoding", "gzip")
	}

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
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

func (handler *WebserverHandler) checkForFallbackFile(logger *zerolog.Logger, w http.ResponseWriter, requestPath string) (file fs.File, requestpath string, finishServing bool) {
	// explicitly requested files do not fall back to index.html, only paths do
	if handler.fallbackFilepath == "" || (path.Ext(requestPath) != "" && path.Ext(requestPath) != ".") {
		http.Error(w, "file not found", http.StatusNotFound)
		return nil, "", true
	}
	requestPath = handler.fallbackFilepath
	file, err := handler.fileSystem.Open(handler.fallbackFilepath)
	if err != nil {
		logger.Error().Err(err).Msg("fallback file not found")
		http.Error(w, "file not found", http.StatusNotFound)
		return nil, "", true
	}
	return file, requestPath, false
}

func (handler *WebserverHandler) getMediaType(requestPath string) string {
	mediaType, ok := handler.config.MediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	return mediaType
}
