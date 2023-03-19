package server

import (
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"strconv"
	"time"
)

// HandlerMiddleware wraps a received handler with another wrapper handler to add functionality
type HandlerMiddleware func(handler http.Handler) http.Handler

// Build a http server from the provided options.
func Build(port int, readTimeout time.Duration, writeTimeout time.Duration, idleTimeout time.Duration,
	handler http.Handler, handlerSetups ...HandlerMiddleware) *http.Server {
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

// CspHeaderReplace replaces the nonce variable in the Content-Security-Header.
func CspHeaderReplace(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return CspHeaderHandler(handler, config.AngularCspReplace.VariableName)
	}
}

// CspFileReplace replaces the nonce variable in all content responses and
// has the hard requirement that a session cookie is present in the context, see server.SessionCookie to add one.
func CspFileReplace(config *Config) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		if config.AngularCspReplace == nil {
			return handler
		}
		return &CspFileHandler{
			Next:         handler,
			VariableName: config.AngularCspReplace.VariableName,
			MediaTypeMap: config.MediaTypeMap,
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

// Fallback adds a fallback route handler.
// THis routes the request to a fallback route on of the given HTTP fallback status codes
func Fallback(fallbackPath string, fallbackCodes ...int) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return FallbackHandler(handler, fallbackPath, fallbackCodes...)
	}
}

// Validate adds to the validate middleware and prevent path transversal attacks by cleaning the request path.
func Validate() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return ValidateHandler(handler)
	}
}

// AccessLog adds an access logging middleware.
func AccessLog() HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return AccessLogHandler(handler)
	}
}

// AccessMetrics collects metrics about bytes send and response status codes and writes
// them to the provided prometheus registerer.
func AccessMetrics(registration *PrometheusRegistration) HandlerMiddleware {
	return func(handler http.Handler) http.Handler {
		return AccessMetricsHandler(handler, registration)
	}
}

// H2C adds a middleware that supports h2c (unencrypted http2)
func H2C(h2cPort int) HandlerMiddleware {
	h2s := &http2.Server{}
	return func(handler http.Handler) http.Handler {
		h2cHandler := h2c.NewHandler(handler, h2s)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Alt-Svc", "h2=\":"+strconv.Itoa(h2cPort)+"\"")
			h2cHandler.ServeHTTP(w, r)
		})
	}
}
