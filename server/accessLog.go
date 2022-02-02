package server

import (
	"net/http"
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

func AccessLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startRaw := r.Context().Value(timerKey)
		var start time.Time
		if startRaw != nil {
			start = startRaw.(time.Time)
		} else {
			start = time.Now()
		}
		logEnter(r.Context(), "access-log")
		metricResponseWriter := &metricResponseWriter{Next: w}
		originalPath := r.URL.String()
		next.ServeHTTP(metricResponseWriter, r)
		logEvent := log.Info()
		requestId := r.Context().Value(requestIdKey)
		if requestId != nil {
			logEvent = logEvent.Str("requestId", requestId.(string))
		}
		logEvent.Dict("httpRequest", zerolog.Dict().
			Str("requestMethod", r.Method).
			Str("requestUrl", originalPath).
			Int("status", metricResponseWriter.StatusCode).
			Int("responseSize", metricResponseWriter.BytesSend).
			Str("userAgent", r.UserAgent()).
			Str("remoteIp", r.RemoteAddr).
			Str("referer", r.Referer()).
			Str("latency", time.Since(start).String())).
			Msg("")
	})
}
