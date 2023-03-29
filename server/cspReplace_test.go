package server_test

import (
	"context"
	"github.com/ngergs/websrv/v3/server"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const path = "dummy_random.js"
const variableName = "123"
const nextHandlerResponse = "test123456"

func TestCspFileReplace(t *testing.T) {
	handler, w, r := getMockedCspFileHandler()
	sessionId := "abc123cde"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	requireReplacedWith(t, sessionId, string(getReceivedData(t, w.Result().Body)))
}

// TestCspFileReplaceSessionMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspFileReplaceSessionMissing(t *testing.T) {
	handler, w, r := getMockedCspFileHandler()
	handler.ServeHTTP(w, r)
	requireReplacedWith(t, "", string(getReceivedData(t, w.Result().Body)))
}

func TestCspHeaderReplace(t *testing.T) {
	handler, w, r := getMockedCspHeaderHandler()
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	sessionId := "321"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	require.Equal(t, "test"+sessionId+"456", w.Result().Header.Get(server.CspHeaderName))
}

// TestCspHeaderReplaceSessionIdMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspHeaderReplaceSessionIdMissing(t *testing.T) {
	handler, w, r := getMockedCspHeaderHandler()
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	handler.ServeHTTP(w, r)
	require.Equal(t, "test456", w.Result().Header.Get(server.CspHeaderName))
}

func requireReplacedWith(t *testing.T, replacedWithExpectation string, replaced string) {
	originalReplaced := strings.ReplaceAll(nextHandlerResponse, variableName, replacedWithExpectation)
	require.Equal(t, originalReplaced, replaced)
}

func getMockedCspFileHandler() (handler *server.CspFileHandler, w *httptest.ResponseRecorder, r *http.Request) {
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(nextHandlerResponse))
		if err != nil {
			log.Error().Msgf("Failed to send response: %v", err)
		}
	}
	handler = &server.CspFileHandler{
		Next:         next,
		VariableName: variableName,
		MediaTypeMap: map[string]string{".js": "application/javascript"},
	}
	r.URL = &url.URL{Path: path}

	return handler, w, r
}

func getMockedCspHeaderHandler() (handler http.Handler, w *httptest.ResponseRecorder, r *http.Request) {
	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(nextHandlerResponse))
		if err != nil {
			log.Error().Msgf("Failed to send response: %v", err)
		}
	}
	handler = server.CspHeaderHandler(next, variableName)
	r.URL = &url.URL{Path: path}

	return handler, w, r
}
