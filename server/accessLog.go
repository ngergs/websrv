package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type metricResponseWriter struct {
	Next       http.ResponseWriter
	StatusCode int
	BytesSend  int
}

func (w *metricResponseWriter) Header() http.Header {
	return w.Next.Header()
}

func (w *metricResponseWriter) Write(data []byte) (int, error) {
	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}
	w.BytesSend += len(data)
	return w.Next.Write(data)
}

func (w *metricResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.Next.WriteHeader(statusCode)
}

// AccessLogHandler returns a http.Handler that adds access-logging on the info level.
func AccessLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startRaw := r.Context().Value(TimerKey)
		var start time.Time
		if startRaw != nil {
			start = startRaw.(time.Time)
		} else {
			start = time.Now()
		}
		logEnter(r.Context(), "access-log")
		metricResponseWriter := &metricResponseWriter{Next: w}
		next.ServeHTTP(metricResponseWriter, r)
		logEvent := log.Info()
		requestId := r.Context().Value(RequestIdKey)
		if requestId != nil {
			logEvent = logEvent.Str("requestId", requestId.(string))
		}

		logEvent.Dict("httpRequest", zerolog.Dict().
			Str("requestMethod", r.Method).
			Str("requestUrl", getFullUrl(r)).
			Int("status", metricResponseWriter.StatusCode).
			Int("responseSize", metricResponseWriter.BytesSend).
			Str("userAgent", r.UserAgent()).
			Str("remoteIp", r.RemoteAddr).
			Str("referer", r.Referer()).
			Str("latency", time.Since(start).String())).
			Msg("")
	})
}

func getFullUrl(r *http.Request) string {
	var sb strings.Builder
	if r.TLS == nil {
		sb.WriteString("http")
	} else {
		sb.WriteString("https")
	}
	sb.WriteString("://")
	sb.WriteString(r.Host)
	if !strings.HasPrefix(r.URL.Path, "/") {
		sb.WriteString("/")
	}
	sb.WriteString(r.URL.Path)
	return sb.String()
}
