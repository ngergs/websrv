package server_test

import (
	"context"
	"github.com/ngergs/websrv/server"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dummyShutdowner struct {
	Ctx          context.Context
	Closed       bool
	ShutdownTime time.Duration
}

func (shutdowner *dummyShutdowner) Shutdown(ctx context.Context) error {
	shutdowner.Ctx = ctx
	time.Sleep(shutdowner.ShutdownTime)
	shutdowner.Closed = true
	return nil
}

func TestGracefulShutdown(t *testing.T) {
	var wg sync.WaitGroup
	shutdowner := &dummyShutdowner{
		Closed:       false,
		ShutdownTime: time.Duration(100) * time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	server.AddGracefulShutdown(ctx, &wg, shutdowner, 0, time.Duration(1)*time.Second)
	assert.False(t, shutdowner.Closed)
	cancel()
	assert.False(t, shutdowner.Closed)
	time.Sleep(shutdowner.ShutdownTime)
	assert.True(t, shutdowner.Closed)
}

func TestGracefulShutdownTimeout(t *testing.T) {
	var wg sync.WaitGroup
	timeoutDuration := time.Duration(1) * time.Second
	shutdowner := &dummyShutdowner{
		Closed:       false,
		ShutdownTime: 10 * timeoutDuration,
	}
	ctx, cancel := context.WithCancel(context.Background())
	server.AddGracefulShutdown(ctx, &wg, shutdowner, 0, timeoutDuration)
	assert.Nil(t, shutdowner.Ctx)
	cancel()
	time.Sleep(time.Duration(100) * time.Millisecond) // wait some time to propagate the cancellation
	assert.NotNil(t, shutdowner.Ctx)
	deadline, ok := shutdowner.Ctx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.Before(time.Now().Add(timeoutDuration)))
}

func TestSigTermCtx(t *testing.T) {
	sigtermCtx := server.SigTermCtx(context.Background())
	assert.False(t, isChannelClosed(sigtermCtx.Done()))
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	assert.True(t, isChannelClosed(sigtermCtx.Done()))
}

func isChannelClosed(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	case <-time.After(time.Duration(100) * time.Millisecond):
		return false
	}
}
