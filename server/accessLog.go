package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var DomainLabel = "domain"
var StatusLabel = "status"

// AccessMetricsHandler collects the bytes send out as well as the status codes as prometheus metrics and writes them
// to the  registry. The registerer has to be prepared via the AccessMetricsRegisterMetrics function.
// registerMetrics usually should be set to true. Setting registerMetrics to false is only for the use case that the same prometheus.Registerer
// should be used for multiple instances of this middleware. Then it should be true only for the first instanced middleware.
func AccessMetricsHandler(next http.Handler, registerer prometheus.Registerer, prometheusNamespace string, registerMetrics bool) http.Handler {
	var bytesSend = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Subsystem: "access",
		Name:      "egress_bytes",
		Help:      "Number of bytes send out from this application.",
	}, []string{DomainLabel})
	var statusCode = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: "access",
		Name:      "http_statuscode",
		Help:      "HTTP Response status code.",
	}, []string{DomainLabel, StatusLabel})
	if registerMetrics {
		registerer.MustRegister(bytesSend, statusCode)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logEnter(r.Context(), "metrics-log")
		metricResponseWriter := &metricResponseWriter{Next: w}
		next.ServeHTTP(metricResponseWriter, r)

		statusCode.With(map[string]string{DomainLabel: r.Host, StatusLabel: strconv.Itoa(metricResponseWriter.StatusCode)}).Inc()
		bytesSend.With(map[string]string{DomainLabel: r.Host}).Add(float64(metricResponseWriter.BytesSend))
	})
}

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
