package server

import (
	"io/fs"
	"net/http"
	"regexp"
	"strconv"
	"text/template"

	"github.com/ngergs/webserver/v2/filesystem"
	"github.com/rs/zerolog/log"
)

// HandlerMiddleware wraps a received handler with another wrapper handler to add functionality
type HandlerMiddleware func(handler http.Handler) http.Handler

// Starts a http server. Blocks till an error occurs.
func Start(name string, port int, errChan chan<- error, handler http.Handler, handlerSetups ...HandlerMiddleware) error {
	//	handler = server.New(filesystem, *fallbackFilepath, config)
	for _, handlerSetup := range handlerSetups {
		handler = handlerSetup(handler)
	}
	fileserver := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: handler,
	}
	log.Info().Msgf("Starting %s server on port %d", name, port)
	return fileserver.ListenAndServe()
}

func Optional(middleware HandlerMiddleware, isActive bool) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if isActive {
			return middleware(handler)
		}
		return handler
	}
}

func Caching(filesystem fs.FS) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		cacheHandler := &CacheHandler{Next: handler, FileSystem: filesystem}
		cacheHandler.Init()
		return cacheHandler
	}
}

func FileReplace(config *Config, filesystem filesystem.ZipFs) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if config.FromHeaderReplace == nil {
			return handler
		}
		return &FileReplaceHandler{
			Next:             handler,
			Filesystem:       filesystem,
			SourceHeaderName: config.FromHeaderReplace.SourceHeaderName,
			FileNamePatter:   regexp.MustCompile(config.FromHeaderReplace.FileNamePattern),
			VariableName:     config.FromHeaderReplace.VariableName,
			Templates:        make(map[string]*template.Template),
			MediaTypeMap:     config.MediaTypeMap,
		}
	}
}

func Header(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return &HeaderHandler{
			Next:    handler,
			Headers: config.Headers,
			Replace: config.FromHeaderReplace,
		}
	}
}

func Gzip(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return GzipHandler(handler, config.GzipMediaTypes)
	}
}

func ValidateClean() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return ValidateCleanHandler(handler)
	}
}

func AccessLog() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return AccessLogHandler(handler)
	}
}

func RequestID() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return RequestIdToCtxHandler(handler)
	}
}

func Timer() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return TimerStartTOCtxHandler(handler)
	}
}
