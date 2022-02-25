package server_test

import (
	"compress/gzip"
	"net/http"
	"testing"

	"github.com/ngergs/websrv/server"
	"github.com/ngergs/websrv/utils"
	"github.com/stretchr/testify/assert"
)

const originalTestMessage = "Just a test message for gzipping"
const gzipCompression = gzip.DefaultCompression

var gzipMediaTypes = []string{"application/javascript"}

func TestGzipCompression(t *testing.T) {
	testZipHandler(t, "gzip", gzipMediaTypes[0], true)
	testZipHandler(t, "gzip", gzipMediaTypes[0]+"t", false)
	testZipHandler(t, "no", gzipMediaTypes[0], false)
	testZipHandler(t, "no", gzipMediaTypes[0]+"t", false)
}

func testZipHandler(t *testing.T, acceptEncoding string, contentType string, expectZipped bool) {
	w, r, next := getDefaultHandlerMocks()
	r.Header.Set("Accept-Encoding", acceptEncoding)
	next.serveHttpFunc = func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(originalTestMessage))
		assert.Nil(t, err)
	}

	gzipHandler := server.GzipHandler(next, gzipCompression, gzipMediaTypes)
	var responseHeader http.Header = map[string][]string{"Content-Type": {contentType}}
	if acceptEncoding == "gzip" {
		w.mock.On("Header").Return(responseHeader)
	}
	gzipHandler.ServeHTTP(w, r)

	w.mock.AssertExpectations(t)
	if expectZipped {
		assert.Equal(t, []string{"gzip"}, responseHeader["Content-Encoding"])
	}
	var expectedResponse []byte
	if expectZipped {
		var err error
		expectedResponse, err = utils.Zip([]byte(originalTestMessage), gzipCompression)
		assert.Nil(t, err)
	} else {
		expectedResponse = []byte(originalTestMessage)

	}
	assert.Equal(t, expectedResponse, w.receivedData.Bytes())

}
