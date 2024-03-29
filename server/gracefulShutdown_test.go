package server_test

import (
	"context"
	"github.com/ngergs/websrv/v3/server"
	"github.com/rs/zerolog/log"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//nolint:containedctx // just for testing
type dummyShutdowner struct {
	mutex        sync.RWMutex
	ctx          context.Context
	closed       bool
	ShutdownTime time.Duration
}

func (shutdowner *dummyShutdowner) Shutdown(ctx context.Context) error {
	shutdowner.mutex.Lock()
	shutdowner.ctx = ctx
	shutdowner.mutex.Unlock()
	time.Sleep(shutdowner.ShutdownTime)
	shutdowner.mutex.Lock()
	defer shutdowner.mutex.Unlock()
	shutdowner.closed = true
	return nil
}

func (shutdowner *dummyShutdowner) isClosed() bool {
	shutdowner.mutex.RLock()
	defer shutdowner.mutex.RUnlock()
	return shutdowner.closed
}

func (shutdowner *dummyShutdowner) getCtx() context.Context {
	shutdowner.mutex.RLock()
	defer shutdowner.mutex.RUnlock()
	return shutdowner.ctx
}

func TestGracefulShutdown(t *testing.T) {
	var wg sync.WaitGroup
	shutdowner := &dummyShutdowner{
		closed:       false,
		ShutdownTime: time.Duration(100) * time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	server.AddGracefulShutdown(ctx, &wg, shutdowner, time.Duration(1)*time.Second)
	require.False(t, shutdowner.isClosed())
	cancel()
	require.False(t, shutdowner.isClosed())
	time.Sleep(2 * shutdowner.ShutdownTime)
	require.True(t, shutdowner.isClosed())
}

func TestGracefulShutdownTimeout(t *testing.T) {
	var wg sync.WaitGroup
	timeoutDuration := time.Duration(1) * time.Second
	shutdowner := &dummyShutdowner{
		closed:       false,
		ShutdownTime: 10 * timeoutDuration,
	}
	ctx, cancel := context.WithCancel(context.Background())
	server.AddGracefulShutdown(ctx, &wg, shutdowner, timeoutDuration)
	require.Nil(t, shutdowner.getCtx())
	cancel()
	time.Sleep(time.Duration(100) * time.Millisecond) // wait some time to propagate the cancellation
	require.NotNil(t, shutdowner.getCtx())
	deadline, ok := shutdowner.getCtx().Deadline()
	require.True(t, ok)
	require.True(t, deadline.Before(time.Now().Add(timeoutDuration)))
}

func TestSigTermCtx(t *testing.T) {
	sigtermCtx := server.SigTermCtx(context.Background(), 0)
	require.False(t, isChannelClosed(sigtermCtx.Done()))
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		log.Err(err).Msg("Sigterm failed")
	}
	require.True(t, isChannelClosed(sigtermCtx.Done()))
}

func isChannelClosed(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	case <-time.After(time.Duration(100) * time.Millisecond):
		return false
	}
}
