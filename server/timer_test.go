package server_test

import (
	"github.com/ngergs/websrv/server"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddingStartTime(t *testing.T) {
	w, r, next := getDefaultHandlerMocks()
	handler := server.TimerStartToCtxHandler(next)
	handler.ServeHTTP(w, r)
	timerStartRaw := next.r.Context().Value(server.TimerKey)
	require.NotNil(t, timerStartRaw)
	timerStart := timerStartRaw.(time.Time)
	// allow some error here as this is set internally when the cookie is created
	expectedStartTime := time.Now()
	require.True(t, timerStart.After(expectedStartTime.Add(-time.Duration(1)*time.Second)))
	require.True(t, timerStart.Before(expectedStartTime.Add(time.Duration(1)*time.Second)))
}
