package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/ngergs/webserver/v2/filesystem"
	"github.com/ngergs/webserver/v2/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type WebserverHandler struct {
	fallbackFilepath string
	fileSystem       filesystem.ZipFs
	config           *Config
	gzipMediaTypes   []string
}

func FileServerHandler(fileSystem filesystem.ZipFs, fallbackFilepath string, config *Config, hasMemoryFs bool) *WebserverHandler {
	handler := &WebserverHandler{
		fallbackFilepath: fallbackFilepath,
		fileSystem:       fileSystem,
		config:           config,
		gzipMediaTypes:   config.GzipMediaTypes,
	}
	if hasMemoryFs {
		handler.gzipMediaTypes = config.GzipMediaTypes
	}
	return handler
}

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logEnter(r.Context(), "webserver")
	logger := log.Ctx(r.Context())
	requestPath := r.URL.Path

	file, requestPath, err := handler.getFileOrFallback(r.Context(), logger, requestPath)
	if err != nil {
		logger.Error().Err(err).Msgf("file %s not found", r.URL.Path)
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer utils.Close(r.Context(), file)

	logger.Debug().Msgf("Serving file %s", requestPath)
	err = handler.setContentHeader(w, requestPath)
	if err != nil {
		logger.Error().Err(err).Msgf("content header error")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeResponse(w, r, file)
}

func (handler *WebserverHandler) getFileOrFallback(ctx context.Context, logger *zerolog.Logger, requestPath string) (file fs.File, requestpath string, err error) {
	file, err = handler.fileSystem.Open(requestPath)
	if err != nil {
		logger.Debug().Err(err).Msgf("file %s not found", requestPath)
		return handler.checkForFallbackFile(logger, requestPath)
	}
	fileInfo, err := file.Stat()
	if fileInfo.IsDir() {
		defer utils.Close(ctx, file)
		return nil, requestPath, fmt.Errorf("requested file is directory")
	}
	return file, requestPath, err
}

func (handler *WebserverHandler) checkForFallbackFile(logger *zerolog.Logger, requestPath string) (file fs.File, requestpath string, err error) {
	// explicitly requested files do not fall back to index.html, only paths do
	if handler.fallbackFilepath == "" || (path.Ext(requestPath) != "" && path.Ext(requestPath) != ".") {
		return nil, "", fmt.Errorf("fallback file not relevant for directories: %s", requestPath)
	}
	requestPath = handler.fallbackFilepath
	file, err = handler.fileSystem.Open(handler.fallbackFilepath)
	if err != nil {
		return nil, "", fmt.Errorf("fallback file %s not found", requestPath)
	}
	return file, requestPath, err
}

func (handler *WebserverHandler) setContentHeader(w http.ResponseWriter, requestPath string) error {
	mediaType, ok := handler.config.MediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mediaType)
	isZipped, err := handler.fileSystem.IsZipped(requestPath)
	if err != nil {
		return fmt.Errorf("failed to resolve whether file is zipped: %s", requestPath)
	}
	if isZipped {
		w.Header().Set("Content-Encoding", "gzip")
	}
	return nil
}

func writeResponse(w http.ResponseWriter, r *http.Request, file io.Reader) {
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, err := io.Copy(w, file)
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
}
