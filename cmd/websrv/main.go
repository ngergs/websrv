package main

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ngergs/websrv/filesystem"
	"github.com/ngergs/websrv/server"
	"github.com/rs/zerolog/log"
)

func main() {
	setup()
	var wg sync.WaitGroup
	sigtermCtx := server.SigTermCtx(context.Background())
	config, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading -config: See server.config.go for the expected structure.")
	}

	unzipfs, zipfs := initFs(config)

	errChan := make(chan error)
	var promRegistration *server.PrometheusRegistration
	if *metrics {
		promRegistration, err = server.AccessMetricsRegister(prometheus.DefaultRegisterer, *metricsNamespace)
		if err != nil {
			log.Error().Err(err).Msg("Could not register custom prometheus metrics.")
		}
	}
	webserver := server.Build(*webServerPort, time.Duration(*readTimeout)*time.Second, time.Duration(*writeTimeout)*time.Second, time.Duration(*idleTimeout)*time.Second,
		server.FileServerHandler(unzipfs, zipfs, *fallbackFilepath, config),
		server.Caching(),
		server.Optional(server.CspReplace(config, unzipfs), config.AngularCspReplace != nil),
		server.Optional(server.Gzip(config, *gzipCompressionLevel), *gzipActive),
		server.Optional(server.SessionId(config), config.AngularCspReplace != nil),
		server.Header(config),
		server.ValidateClean(),
		server.Optional(server.AccessMetrics(promRegistration), *metrics),
		server.Optional(server.AccessLog(), *accessLog),
		server.RequestID(),
		server.Timer())
	log.Info().Msgf("Starting webserver server on port %d", *webServerPort)
	srvCtx := context.WithValue(sigtermCtx, server.ServerName, "file server")
	server.AddGracefulShutdown(srvCtx, &wg, webserver, time.Duration(*shutdownDelay)*time.Second, time.Duration(*shutdownTimeout)*time.Second)
	go func() { errChan <- webserver.ListenAndServe() }()

	if *metrics {
		metricsServer := server.Build(*metricsPort, time.Duration(*readTimeout)*time.Second, time.Duration(*writeTimeout)*time.Second, time.Duration(*idleTimeout)*time.Second,
			promhttp.Handler(), server.Optional(server.AccessLog(), *metricsAccessLog))
		metricsCtx := context.WithValue(sigtermCtx, server.ServerName, "prometheus metrics server")
		server.AddGracefulShutdown(metricsCtx, &wg, metricsServer, time.Duration(*shutdownDelay)*time.Second, time.Duration(*shutdownTimeout)*time.Second)
		go func() {
			log.Info().Msgf("Listening for prometheus metric scrapes under container port tcp/%s", metricsServer.Addr[1:])
			errChan <- metricsServer.ListenAndServe()
		}()
	}

	go logErrors(errChan)

	// stop health server after everything else has stopped
	if *health {
		healthServer := server.Build(*healthPort, time.Duration(*readTimeout)*time.Second, time.Duration(*writeTimeout)*time.Second, time.Duration(*idleTimeout)*time.Second,
			server.HealthCheckHandler(),
			server.Optional(server.AccessLog(), *healthAccessLog),
		)
		log.Info().Msgf("Starting healthcheck server on port %d", *healthPort)
		healthCtx := context.WithValue(context.Background(), server.ServerName, "health server")
		// 1 second is sufficient for health checks to shut down
		server.RunTillWaitGroupFinishes(healthCtx, &wg, healthServer, errChan, time.Duration(1)*time.Second)
	} else {
		wg.Wait()
	}
}

// initFs loads the non-zipped and zipped fs according to the config
func initFs(config *server.Config) (unzipfs fs.ReadFileFS, zipfs fs.ReadFileFS) {
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		memoryFs, err := filesystem.NewMemoryFs(targetDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing read-only filesystem.")
		}
		unzipfs = memoryFs
		if *gzipActive {
			log.Debug().Msg("Zipping in memory filesystem")
			zipfs, err = memoryFs.Zip(config.GzipFileExtensions())
			if err != nil {
				log.Fatal().Err(err).Msg("Error preparing zipped read-only filesystem.")
			}
		}
	} else {
		log.Info().Msg("Using the os filesystem")
		unzipfs = &filesystem.ReadFileFS{FS: os.DirFS(targetDir)}
	}
	return
}

// logErrors listens to the provided errChan and logs the received errors
func logErrors(errChan <-chan error) {
	for err := range errChan {
		if errors.Is(err, http.ErrServerClosed) {
			// thrown from listen, serve and listenAndServe during graceful shutdown
			log.Debug().Err(err).Msg("Expected graceful shutdown error")
		} else {
			log.Fatal().Err(err).Msg("Error from server: %v")
		}
	}
}
