package server

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

var DomainLabel = "domain"
var StatusLabel = "status"

// PrometheusRegistration wraps a prometheus registerer and corresponding registered types.
type PrometheusRegistration struct {
	bytesSend  *prometheus.CounterVec
	statusCode *prometheus.CounterVec
}

// AccessMetricsRegister registrates the relevant prometheus types and returns a custom registration type
func AccessMetricsRegister(registerer prometheus.Registerer, prometheusNamespace string) (*PrometheusRegistration, error) {
	var bytesSend = prometheus.NewCounterVec(prometheus.CounterOpts{
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

	err := registerer.Register(bytesSend)
	if err != nil {
		return nil, fmt.Errorf("failed to register egress_bytes metric: %w", err)
	}
	err = registerer.Register(statusCode)
	if err != nil {
		return nil, fmt.Errorf("failed to register http_statuscode metric: %w", err)
	}
	return &PrometheusRegistration{
		bytesSend:  bytesSend,
		statusCode: statusCode,
	}, nil
}

// AccessMetricsHandler collects the bytes send out as well as the status codes as prometheus metrics and writes them
// to the  registry. The registerer has to be prepared via the AccessMetricsRegister function.
func AccessMetricsHandler(next http.Handler, registration *PrometheusRegistration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		registration.statusCode.With(map[string]string{DomainLabel: r.Host, StatusLabel: strconv.Itoa(m.Code)}).Inc()
		registration.bytesSend.With(map[string]string{DomainLabel: r.Host}).Add(float64(m.Written))
	})
}

// AccessLogHandler returns a http.Handler that adds access-logging on the info level.
//
//nolint:zerologlint // linter does not understand that we dispatch logEvent later on
func AccessLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		logEvent := log.Info()
		requestId := r.Context().Value(middleware.RequestIDKey)
		if requestId != nil {
			if requestIdStr, ok := requestId.(string); ok {
				logEvent = logEvent.Str("requestId", requestIdStr)
			} else {
				log.Warn().Msgf("Request id is not, but not a string value: %v", requestId)
			}
		}
		logEvent.Dict("httpRequest", zerolog.Dict().
			Str("requestMethod", r.Method).
			Str("requestUrl", getFullUrl(r)).
			Int("status", m.Code).
			Str("responseSize", strconv.FormatInt(m.Written, 10)).
			Str("userAgent", r.UserAgent()).
			Str("remoteIp", r.RemoteAddr).
			Str("referer", r.Referer()).
			Str("latency", fmt.Sprintf("%.09fs", m.Duration.Seconds()))).
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
