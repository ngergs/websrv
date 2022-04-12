package main

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

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
	webserver := server.Build(*webServerPort,
		server.FileServerHandler(unzipfs, zipfs, *fallbackFilepath, config),
		server.Caching(unzipfs),
		server.Optional(server.CspReplace(config, unzipfs), config.AngularCspReplace != nil),
		server.Optional(server.Gzip(config, *gzipCompressionLevel), *gzipActive),
		server.Optional(server.SessionId(config), config.AngularCspReplace != nil),
		server.Header(config),
		server.ValidateClean(),
		server.Optional(server.AccessLog(), *accessLog),
		server.RequestID(),
		server.Timer())
	log.Info().Msgf("Starting webserver server on port %d", *webServerPort)
	srvCtx := context.WithValue(sigtermCtx, server.ServerName, "file server")
	server.AddGracefulShutdown(srvCtx, &wg, webserver, time.Duration(*shutdownDelay)*time.Second, time.Duration(*shutdownTimeout)*time.Second)
	go func() { errChan <- webserver.ListenAndServe() }()

	if *health {
		healthServer := server.Build(*healthPort,
			server.HealthCheckHandler(),
			server.Optional(server.AccessLog(), *healthAccessLog),
		)
		log.Info().Msgf("Starting healthcheck server on port %d", *healthPort)
		healthCtx := context.WithValue(sigtermCtx, server.ServerName, "health server")
		server.AddGracefulShutdown(healthCtx, &wg, healthServer, time.Duration(*shutdownDelay)*time.Second, time.Duration(*shutdownTimeout)*time.Second)
		go func() { errChan <- healthServer.ListenAndServe() }()
	}

	go logErrors(errChan)
	wg.Wait()
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
