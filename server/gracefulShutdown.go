package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

var ServerName = &ContextKey{val: "serverName"}

// Shutdowner are functions that support a Shutdown operation. It is the responsibility of the interface implementer to honor the context deadline.
type Shutdowner interface {
	Shutdown(context.Context) error
}

// AddGracefulShutdown intercepts the cancel function of the received ctx and calls the shutdowner.Shutdown interface instead.
// if timeout is not null a context with a deadline is prepared prior to the Shutdown call.
// It is the responsibility of the Shutdowner interface implementer to honor this context deadline.
// The waitgroup is incremented by one immediately and one is released when the shutdown has finished.
func AddGracefulShutdown(ctx context.Context, wg *sync.WaitGroup, shutdowner Shutdowner, shutdownDelay time.Duration, timeout time.Duration) {
	wg.Add(1)
	go func() {
		<-ctx.Done()
		logShutdown(ctx, shutdownDelay, timeout)
		time.Sleep(shutdownDelay)
		shutdownCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
		defer cancel()
		err := shutdowner.Shutdown(shutdownCtx)
		wg.Done()
		if err != nil {
			log.Warn().Err(err).Msg("Error during graceful shutdown")
		}
	}()
}

// RunTillWaitGroupFinishes runs the server argument until the WaitGroup wg finishes.
// Subsequently, a graceful shutdown with the given timeout argument is executed.
// Blocks till then.
func RunTillWaitGroupFinishes(ctx context.Context, wg *sync.WaitGroup, server *http.Server, errChan chan<- error, timeout time.Duration) {
	go func() { errChan <- server.ListenAndServe() }()
	wg.Wait()
	logShutdown(ctx, 0, timeout)
	shutdownCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		errChan <- err
	}
}

// logShutdown logs the relevant info for the shutdown and extracts the optional server name from the context
func logShutdown(ctx context.Context, shutdownDelay time.Duration, timeout time.Duration) {
	serverName := ctx.Value(ServerName)
	if serverName != nil {
		log.Info().Msgf("%s: Graceful shutdown with delay %.0fs and timeout %.0fs", serverName, shutdownDelay.Seconds(), timeout.Seconds())
	} else {
		log.Info().Msgf("Graceful shutdown with delay %.0fs and timeout %.0fs", shutdownDelay.Seconds(), timeout.Seconds())
	}
}

// SigTermCtx intercepts the syscall.SIGTERM and returns the information in the form of a wrapped context whose cancel function is called when the SIGTERM signal is received.
func SigTermCtx(ctx context.Context) context.Context {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sigterm := <-termChan
		log.Info().Msgf("Received system call: %v", sigterm)
		cancel()
	}()
	return ctx
}
