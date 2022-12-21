package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"io/fs"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// HandlerMiddleware wraps a received handler with another wrapper handler to add functionality
type HandlerMiddleware func(handler http.Handler) http.Handler

// Build a http server from the provided options.
func Build(port int, readTimeout time.Duration, writeTimeout time.Duration, idleTimeout time.Duration,
	handler http.Handler, handlerSetups ...HandlerMiddleware) *http.Server {
	//	handler = server.New(filesystem, *fallbackFilepath, config)
	for _, handlerSetup := range handlerSetups {
		handler = handlerSetup(handler)
	}
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return server
}

// Optional sets the middleware if the isActive condition is fulfilled
func Optional(middleware HandlerMiddleware, isActive bool) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if isActive {
			return middleware(handler)
		}
		return handler
	}
}

// Caching adds a caching middleware handler which uses the ETag HTTP response and If-None-Match HTTP request headers.
// This requires that all following handler only serve static resources. Following handlers will only be called when a cache mismatch occurs.
func Caching() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return NewCacheHandler(handler)
	}
}

// CspReplace has the hard requirement that a session cookie is present in the context, see server.SessionCookie to add one.
func CspReplace(config *Config, filesystem fs.ReadFileFS) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if config.AngularCspReplace == nil {
			return handler
		}
		return &CspReplaceHandler{
			Next:           handler,
			Filesystem:     filesystem,
			FileNamePatter: regexp.MustCompile(config.AngularCspReplace.FileNamePattern),
			VariableName:   config.AngularCspReplace.VariableName,
			MediaTypeMap:   config.MediaTypeMap,
		}
	}
}

// SessionId adds a session cookie adding middleware
func SessionId(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return SessionCookieHandler(handler, config.AngularCspReplace.CookieName, time.Duration(config.AngularCspReplace.CookieMaxAge)*time.Second)
	}
}

// Header adds a static HTTP header adding middleware
func Header(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return &HeaderHandler{
			Next:    handler,
			Headers: config.Headers,
		}
	}
}

// Gzip adds an on-demand gzipping middleware.
// Gzip is only applied when the Accept-Encoding: gzip HTTP request header is present
// and the Content-Type of the response matches the config options.
func Gzip(config *Config, compressionLevel int) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return GzipHandler(handler, compressionLevel, config.GzipMediaTypes)
	}
}

// ValidateClean adds to the validate middleware and prevent path transversal attacks by cleaning the request path.
func ValidateClean() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return ValidateCleanHandler(handler)
	}
}

// AccessLog adds an access logging middleware.
func AccessLog() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return AccessLogHandler(handler)
	}
}

// AccessMetrics collects metrics about bytes send and response status codes and writes
// them to the provided prometheus registerer
func AccessMetrics(registerer prometheus.Registerer, namespace string) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return AccessMetricsHandler(handler, registerer, namespace)
	}
}

// RequestID adds a middleware that adds a randomly generated request id to the request context.
func RequestID() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return RequestIdToCtxHandler(handler)
	}
}

// Timer adds a middleware that adds a started timer to the request context for time measuring purposes.
func Timer() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return TimerStartToCtxHandler(handler)
	}
}
