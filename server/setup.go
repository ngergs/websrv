package server

import (
	"io/fs"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ngergs/webserver/filesystem"
)

// HandlerMiddleware wraps a received handler with another wrapper handler to add functionality
type HandlerMiddleware func(handler http.Handler) http.Handler

// Builds a http server from the provided options.
func Build(port int, handler http.Handler, handlerSetups ...HandlerMiddleware) *http.Server {
	//	handler = server.New(filesystem, *fallbackFilepath, config)
	for _, handlerSetup := range handlerSetups {
		handler = handlerSetup(handler)
	}
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: handler,
	}
	return server
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

// CspReplace has the hard requirement that a session cookie is present in the context, see server.SessionCookie to add one.
func CspReplace(config *Config, filesystem filesystem.ZipFs) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if config.AngularCspReplace == nil {
			return handler
		}
		return &CspReplaceHandler{
			Next:           handler,
			Filesystem:     filesystem,
			FileNamePatter: regexp.MustCompile(config.AngularCspReplace.FileNamePattern),
			VariableName:   config.AngularCspReplace.VariableName,
			Replacer:       make(map[string]*replacerCollection),
			MediaTypeMap:   config.MediaTypeMap,
		}
	}
}

func SessionId(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return SessionCookieHandler(handler, config.AngularCspReplace.CookieName, time.Duration(config.AngularCspReplace.CookieMaxAge)*time.Second)
	}
}

func Header(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return &HeaderHandler{
			Next:    handler,
			Headers: config.Headers,
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
