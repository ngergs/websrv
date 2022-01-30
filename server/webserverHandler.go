package server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/rs/zerolog/log"
)

// needed due to https://github.com/golang/go/issues/32350
var mediaTypeMapping map[string]string = map[string]string{
	".js":   "application/javascript",
	".css":  "text/css",
	".html": "text/html; charset=UTF-8",
	".jpg":  "image/jpeg",
	".avif": "image/avif",
	".jxl":  "image/jxl",
}

type WebserverHandler struct {
	FallbackFilepath string
	FileSystem       fs.FS
	HttpheaderConfig *HttpHeaderConfig
}

func (handler *WebserverHandler) getMediaType(requestPath string) string {
	mediaType, ok := mediaTypeMapping[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	return mediaType
}

func (handler *WebserverHandler) tryGetFile(requestPath string) (fs.File, error) {
	file, err := handler.FileSystem.Open(requestPath)
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

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "This server only supports GET", http.StatusMethodNotAllowed)
		return
	}
	// remove leading / from path to make it relative
	// important to do this after cleaning, else relative paths may remain
	if !path.IsAbs(r.URL.Path) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	requestPath := path.Clean(r.URL.Path)[1:]

	file, err := handler.tryGetFile(requestPath)
	if err != nil {
		log.Debug().Err(err).Msgf("file %s not found", requestPath)
		// requested files do not fall back to index.html
		if handler.FallbackFilepath == "" || (path.Ext(requestPath) != "" && path.Ext(requestPath) != ".") {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		file, err = handler.FileSystem.Open(handler.FallbackFilepath)
		requestPath = handler.FallbackFilepath
		if err != nil {
			log.Error().Err(err).Msg("fallback file not found")
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
	}
	defer file.Close()

	// needed due to https://github.com/golang/go/issues/32350
	w.Header()["Content-Type"] = []string{handler.getMediaType(requestPath)}
	if handler.HttpheaderConfig != nil {
		for k, v := range handler.HttpheaderConfig.Headers {
			w.Header()[k] = v
		}
	}
	_, err = io.Copy(w, file)
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
}
