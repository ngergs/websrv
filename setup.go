package main

import (
	"io/fs"
	"net/http"
	"regexp"
	"text/template"

	"github.com/ngergs/webserver/v2/server"
)

// HandlerMiddleware wraps a received handler with another wrapper handler to add functionality
type HandlerMiddleware func(handler http.Handler) http.Handler

func Caching(filesystem fs.FS) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		cacheHandler := &server.CacheHandler{Next: handler, FileSystem: filesystem}
		cacheHandler.Init()
		return cacheHandler
	}
}

func FileReplace(config *server.Config, filesystem fs.FS) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if config.FromHeaderReplace == nil {
			return handler
		}
		return &server.FileReplaceHandler{
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

func Header(config *server.Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return &server.HeaderHandler{
			Next:    handler,
			Headers: config.Headers,
			Replace: config.FromHeaderReplace,
		}
	}
}

func Gzip(config *server.Config, active bool) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if !active {
			return handler
		}
		return server.GzipHandler(handler, config.GzipMediaTypes)
	}
}

func ValidateClean() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return server.ValidateCleanHandler(handler)
	}
}

func AccessLog(active bool) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if !active {
			return handler
		}
		return server.AccessLogHandler(handler)
	}
}

func RequestID() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return server.RequestIdToCtxHandler(handler)
	}
}
