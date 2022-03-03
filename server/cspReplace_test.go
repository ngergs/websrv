package server_test

import (
	"context"
	"io/fs"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/server"
	"github.com/stretchr/testify/assert"
)

const path = "dummy_random.js"
const variableName = "123"
const nextHandlerResponse = "just a random string for the next handler response"

func TestCspFileReplace(t *testing.T) {
	handler, fs, w, r, _ := getMockedCspHandler(t)
	sessionId := "abc123cde"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assertReplacedWith(t, fs, sessionId, w.receivedData.String())
}

// TestCspFileReplaceSessionMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspFileReplaceSessionMissing(t *testing.T) {
	handler, fs, w, r, _ := getMockedCspHandler(t)
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assertReplacedWith(t, fs, "", w.receivedData.String())
}

//TestCspFileReplacFilePatternMissmatch checks that the next handler is called when the file patterns do not match
func TestCspFileReplacFilePatternMissmatch(t *testing.T) {
	handler, _, w, r, _ := getMockedCspHandler(t)
	handler.FileNamePatter = regexp.MustCompile("^$")
	sessionId := "abc123cde"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assert.Equal(t, nextHandlerResponse, w.receivedData.String())
}

func TestCspHeaderReplace(t *testing.T) {
	handler, _, w, r, responseHeader := getMockedCspHandler(t)
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	sessionId := "321"
	r = r.WithContext(context.WithValue(context.Background(), server.SessionIdKey, sessionId))
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assert.Equal(t, "test"+sessionId+"456", responseHeader.Get(server.CspHeaderName))
}

// TestCspHeaderReplaceSessionIdMissing tests that the VariableName is replaced with "" if the sessionID is absent
func TestCspHeaderReplaceSessionIdMissing(t *testing.T) {
	handler, _, w, r, responseHeader := getMockedCspHandler(t)
	w.Header().Set(server.CspHeaderName, "test"+variableName+"456")
	handler.ServeHTTP(w, r)
	w.mock.AssertExpectations(t)
	assert.Equal(t, "test456", responseHeader.Get(server.CspHeaderName))
}

func assertReplacedWith(t *testing.T, fs fs.ReadFileFS, replacedWithExpectation string, replaced string) {
	original, err := fs.ReadFile(path)
	assert.Nil(t, err)
	originalReplaced := strings.ReplaceAll(string(original), variableName, replacedWithExpectation)
	assert.Equal(t, originalReplaced, replaced)
}

func getMockedCspHandler(t *testing.T) (handler *server.CspReplaceHandler, fs fs.ReadFileFS, w *mockResponseWriter, r *http.Request, responseHeader http.Header) {
	fs, err := filesystem.NewMemoryFs("../benchmark")
	assert.Nil(t, err)

	w, r, next := getDefaultHandlerMocks()
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(nextHandlerResponse))
	}
	handler = &server.CspReplaceHandler{
		Next:           next,
		Filesystem:     fs,
		FileNamePatter: regexp.MustCompile(".*"),
		VariableName:   variableName,
		MediaTypeMap:   map[string]string{".js": "application/javascript"},
	}
	r.URL = &url.URL{Path: path}
	responseHeader = make(map[string][]string)
	w.mock.On("Header").Return(responseHeader)

	return handler, fs, w, r, responseHeader
}
