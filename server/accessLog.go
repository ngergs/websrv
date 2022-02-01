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
		log.Ctx(r.Context()).Debug().Msg("Entering access log handler")
		start := time.Now()
		metricResponseWriter := &metricResponseWriter{Next: w}
		originalPath := r.URL.String()
		next.ServeHTTP(metricResponseWriter, r)
		log.Info().
			Str("requestId", r.Context().Value(requestIdKey).(string)).
			Dict("httpRequest", zerolog.Dict().
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
