package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/ngergs/webserver/utils"
	"github.com/rs/zerolog/log"
)

type WebserverHandler struct {
	fetch              fileFetch
	mediaTypeMap       map[string]string
	gzipMediaTypes     []string
	gzipFileExtensions []string
}

type fileFetch func(ctx context.Context, requestPath string, zipOk bool) (file fs.File, path string, zipped bool, err error)

func fetchWrapper(nextZip fileFetch, nextUnzip fileFetch) fileFetch {
	return func(ctx context.Context, path string, zipOk bool) (fs.File, string, bool, error) {
		if zipOk && nextZip != nil {
			return nextZip(ctx, path, zipOk)
		}
		return nextUnzip(ctx, path, zipOk)
	}
}

func fsFetch(name string, filesystem fs.FS, zipped bool, next fileFetch) fileFetch {
	if filesystem == nil {
		log.Debug().Msgf("filesystem absent for fetcher %s, skipping configuration", name)
		return nil
	}
	return func(ctx context.Context, path string, zipOk bool) (fs.File, string, bool, error) {
		file, err := filesystem.Open(path)
		log.Ctx(ctx).Debug().Msgf("fileFetch %s file miss for %s, trying next", name, path)
		if err != nil {
			if next != nil {
				return next(ctx, path, zipOk)
			}
		}
		return file, path, zipped, err
	}
}

func fallbackFetch(name string, filesystem fs.FS, zipped bool, fallbackFilepath string) fileFetch {
	if filesystem == nil {
		log.Debug().Msgf("filesystem absent for fetcher %s, skipping configuration", name)
		return nil
	}
	return func(ctx context.Context, path string, zipOk bool) (fs.File, string, bool, error) {
		if zipped && !zipOk {
			return nil, "", false, fmt.Errorf("requested unzipped file %s from zipped fallback", name)
		}
		file, err := filesystem.Open(fallbackFilepath)
		return file, fallbackFilepath, zipped, err
	}
}

// FileServerHandler implements the actual fileserver logic. zipfs can be set to nil if no pre-zipped file have been prepared.
func FileServerHandler(fs fs.FS, zipfs fs.FS, fallbackFilepath string, config *Config) *WebserverHandler {
	fallbackFetcher := fetchWrapper(fallbackFetch("zipped", zipfs, true, fallbackFilepath),
		fallbackFetch("unzipped", fs, false, fallbackFilepath))
	fetcher := fetchWrapper(fsFetch("zipped", zipfs, true, fallbackFetcher),
		fsFetch("unzipped", fs, false, fallbackFetcher))
	handler := &WebserverHandler{
		fetch:              fetcher,
		mediaTypeMap:       config.MediaTypeMap,
		gzipMediaTypes:     config.GzipMediaTypes,
		gzipFileExtensions: config.GzipFileExtensions(),
	}
	return handler
}

func (handler *WebserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logEnter(ctx, "webserver")
	logger := log.Ctx(ctx)

	zipOk := utils.ContainsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip") &&
		utils.Contains(handler.gzipFileExtensions, path.Ext(r.URL.Path))
	file, servedPath, zipped, err := handler.fetch(ctx, r.URL.Path, zipOk)
	if err != nil {
		logger.Error().Err(err).Msgf("file %s not found", r.URL.Path)
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer utils.Close(ctx, file)

	logger.Debug().Msgf("Serving file %s", servedPath)
	err = handler.setContentHeader(w, servedPath, zipped)
	if err != nil {
		logger.Error().Err(err).Msgf("content header error")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeResponse(w, r, file)
}

func (handler *WebserverHandler) setContentHeader(w http.ResponseWriter, requestPath string, zipped bool) error {
	mediaType, ok := handler.mediaTypeMap[path.Ext(requestPath)]
	if !ok {
		mediaType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mediaType)
	if zipped {
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
