package server_test

import (
	"compress/gzip"
	"github.com/ngergs/websrv/server"
	"net/http"
	"testing"

	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/require"
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
		w.Header().Set("Content-Type", contentType)
		_, err := w.Write([]byte(originalTestMessage))
		require.Nil(t, err)
	}

	gzipHandler := server.GzipHandler(next, gzipCompression, gzipMediaTypes)
	gzipHandler.ServeHTTP(w, r)

	if expectZipped {
		require.Equal(t, "gzip", w.Result().Header.Get("Content-Encoding"))
	}
	var expectedResponse []byte
	if expectZipped {
		var err error
		expectedResponse, err = utils.Zip([]byte(originalTestMessage), gzipCompression)
		require.Nil(t, err)
	} else {
		expectedResponse = []byte(originalTestMessage)

	}
	require.Equal(t, expectedResponse, getReceivedData(t, w.Result().Body))

}
