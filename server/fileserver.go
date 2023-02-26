package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/ngergs/websrv/internal/utils"
	"github.com/rs/zerolog/log"
)

// fileserverHandler is the main struct used for serving files
type fileserverHandler struct {
	fetch        fileFetch
	mediaTypeMap map[string]string
}

// fileFetch is a named function type. It is used to build chains of filefetching logics. The main route is zip/raw -> fallback -> zip/raw. See FileServerHandler for the setup.
type fileFetch func(ctx context.Context, requestPath string, zipAccept bool) (file fs.File, path string, zipped bool, err error)

// fetchWrapperZipOrNot calls the first provided fileFetch argument when the content should be zipped and the second one for raw content dependent on the Accept-Encoding header and the file extension.
func fetchWrapperZipOrNot(gzipFileExtensions []string, nextZip fileFetch, nextUnzip fileFetch) fileFetch {
	if nextZip == nil {
		return func(ctx context.Context, filepath string, zipAccept bool) (fs.File, string, bool, error) {
			return nextUnzip(ctx, filepath, zipAccept)
		}
	}
	return func(ctx context.Context, filepath string, zipAccept bool) (fs.File, string, bool, error) {
		if zipAccept && utils.Contains(gzipFileExtensions, path.Ext(filepath)) {
			return nextZip(ctx, filepath, zipAccept)
		}
		return nextUnzip(ctx, filepath, zipAccept)
	}
}

// fallbackFetch sets the path to the fallbackPath and calls the next fileFetch
func fallbackFetch(fallbackPath string, next fileFetch) fileFetch {
	return func(ctx context.Context, path string, zipAccept bool) (fs.File, string, bool, error) {
		return next(ctx, fallbackPath, zipAccept)
	}
}

// fsFetch holds the actual logic for opening the file from the filesystem. Calls the next filefetcher if it is non-nill and an error occurs.
func fsFetch(name string, filesystem fs.FS, zipped bool, next fileFetch) fileFetch {
	if filesystem == nil {
		log.Debug().Msgf("filesystem absent for fetcher %s, skipping configuration", name)
		return nil
	}
	return func(ctx context.Context, path string, zipAccept bool) (fs.File, string, bool, error) {
		logger := log.Ctx(ctx)
		file, err := filesystem.Open(path)
		logger.Debug().Msgf("fileFetch %s file miss for %s, trying next", name, path)
		if err != nil {
			logger.Debug().Err(err).Msg("fsFetch missmatch")
			if next != nil {
				return next(ctx, path, zipAccept)
			}
			return file, path, zipped, err
		}

		stat, err := file.Stat()
		if err != nil || stat.IsDir() {
			if err != nil {
				log.Error().Msgf("Could not get file stats: %v", err)
			}
			if next != nil {
				return next(ctx, path, zipAccept)
			}
		}
		return file, path, zipped, err
	}
}

// FileServerHandler implements the actual fileserver logic. zipfs can be set to nil if no pre-zipped file have been prepared.
func FileServerHandler(fs fs.FS, zipfs fs.FS, fallbackFilepath string, config *Config) http.Handler {
	fallbackFetcher := fallbackFetch(fallbackFilepath,
		fetchWrapperZipOrNot(config.GzipFileExtensions(),
			fsFetch("zipped", zipfs, true, nil),
			fsFetch("unzipped", fs, false, nil)))
	fetcher := fetchWrapperZipOrNot(config.GzipFileExtensions(),
		fsFetch("zipped", zipfs, true, fallbackFetcher),
		fsFetch("unzipped", fs, false, fallbackFetcher))
	handler := &fileserverHandler{
		fetch:        fetcher,
		mediaTypeMap: config.MediaTypeMap,
	}
	return handler
}

// ServeHttp implements the http.Handler interface
func (handler *fileserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logEnter(ctx, "webserver")
	logger := log.Ctx(ctx)

	zipAccept := utils.ContainsAfterSplit(r.Header.Values("Accept-Encoding"), ",", "gzip")

	file, servedPath, zipped, err := handler.fetch(ctx, r.URL.Path, zipAccept)
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

// setContentHeader sets the Content-Type and the Content-Encoding (gzip or absent).
func (handler *fileserverHandler) setContentHeader(w http.ResponseWriter, requestPath string, zipped bool) error {
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

// writeResponse just streams the file content to the writer w and handles errors.
func writeResponse(w http.ResponseWriter, r *http.Request, file io.Reader) {
	if r.Method == http.MethodHead {
		return
	}
	_, err := io.Copy(w, file)
	if err != nil {
		log.Warn().Err(err).Msg("error copying requested file")
		http.Error(w, "failed to copy requested file, you can retry.", http.StatusInternalServerError)
		return
	}
}
