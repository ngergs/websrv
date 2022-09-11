package server_test

import (
	"context"
	"github.com/ngergs/websrv/server"
	"github.com/rs/zerolog/log"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/stretchr/testify/assert"
)

const path = "dummy_random.js"
const variableName = "123"
const nextHandlerResponse = "just a random string for the next handler response"

func TestCspFileReplace(t *testing.T) {
	handler, filesystem, w, r := getMockedCspHandler(t)
	sessionId := "abc123cde"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	assertReplacedWith(t, filesystem, sessionId, string(getReceivedData(t, w.Result().Body)))
}

// TestCspFileReplaceSessionMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspFileReplaceSessionMissing(t *testing.T) {
	handler, filesystem, w, r := getMockedCspHandler(t)
	handler.ServeHTTP(w, r)
	assertReplacedWith(t, filesystem, "", string(getReceivedData(t, w.Result().Body)))
}

// TestCspFileReplacFilePatternMissmatch checks that the next handler is called when the file patterns do not match
func TestCspFileReplacFilePatternMissmatch(t *testing.T) {
	handler, _, w, r := getMockedCspHandler(t)
	handler.FileNamePatter = regexp.MustCompile("^$")
	sessionId := "abc123cde"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	assert.Equal(t, nextHandlerResponse, string(getReceivedData(t, w.Result().Body)))
}

func TestCspHeaderReplace(t *testing.T) {
	handler, _, w, r := getMockedCspHandler(t)
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	sessionId := "321"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	assert.Equal(t, "test"+sessionId+"456", w.Result().Header.Get(server.CspHeaderName))
}

// TestCspHeaderReplaceSessionIdMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspHeaderReplaceSessionIdMissing(t *testing.T) {
	handler, _, w, r := getMockedCspHandler(t)
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	handler.ServeHTTP(w, r)
	assert.Equal(t, "test456", w.Result().Header.Get(server.CspHeaderName))
}

func assertReplacedWith(t *testing.T, fs fs.ReadFileFS, replacedWithExpectation string, replaced string) {
	original, err := fs.ReadFile(path)
	assert.Nil(t, err)
	originalReplaced := strings.ReplaceAll(string(original), variableName, replacedWithExpectation)
	assert.Equal(t, originalReplaced, replaced)
}

func getMockedCspHandler(t *testing.T) (handler *server.CspReplaceHandler, fs fs.ReadFileFS, w *httptest.ResponseRecorder, r *http.Request) {
	filesystem, err := filesystem.NewMemoryFs("../test/benchmark")
	assert.Nil(t, err)

	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(nextHandlerResponse))
		if err != nil {
			log.Error().Msgf("Failed to send response: %v", err)
		}
	}
	handler = &server.CspReplaceHandler{
		Next:           next,
		Filesystem:     filesystem,
		FileNamePatter: regexp.MustCompile(".*"),
		VariableName:   variableName,
		MediaTypeMap:   map[string]string{".js": "application/javascript"},
	}
	r.URL = &url.URL{Path: path}

	return handler, filesystem, w, r
}
