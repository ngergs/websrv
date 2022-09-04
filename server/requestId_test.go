package server_test

import (
	"github.com/ngergs/websrv/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddingRequestId(t *testing.T) {
	requestId1 := serveAndGetRequestId(t)
	requestId2 := serveAndGetRequestId(t)
	assert.NotEqual(t, requestId1, requestId2)

}

func serveAndGetRequestId(t *testing.T) string {
	w, r, next := getDefaultHandlerMocks()
	handler := server.RequestIdToCtxHandler(next)
	handler.ServeHTTP(w, r)
	requestIdRaw := next.r.Context().Value(server.RequestIdKey)
	assert.NotNil(t, requestIdRaw)
	requestId := requestIdRaw.(string)
	assert.NotEqual(t, "", requestId)
	return requestId
}
