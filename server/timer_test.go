package server_test

import (
	"github.com/ngergs/websrv/server"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddingStartTime(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.TimerStartToCtxHandler(next)
	handler.ServeHTTP(w, r)
	timerStartRaw := next.r.Context().Value(server.TimerKey)
	assert.NotNil(t, timerStartRaw)
	timerStart := timerStartRaw.(time.Time)
	// allow some error here as this is set internally when the cookie is created
	expectedStartTime := time.Now()
	assert.True(t, timerStart.After(expectedStartTime.Add(-time.Duration(1)*time.Second)))
	assert.True(t, timerStart.Before(expectedStartTime.Add(time.Duration(1)*time.Second)))
}
