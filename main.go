package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ngergs/webserver/v2/filesystems"
	"github.com/ngergs/webserver/v2/server"
	"github.com/rs/zerolog/log"
)

func startFileServer(config *server.Config, errChan chan<- error) {
	flag.Parse()
	var filesystem fs.FS
	var err error
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		filesystem, err = filesystems.NewMemoryFs(targetDir)
		if err != nil {
			errChan <- fmt.Errorf("error initializing in-memory-fs: %w", err)
			return
		}
	} else {
		log.Info().Msg("Using the os filesystem")
		filesystem = os.DirFS(targetDir)
	}
	var handler http.Handler
	handler, err = server.New(filesystem, *fallbackFilepath, config)
	if err != nil {
		errChan <- fmt.Errorf("error initializing webserver handler: %w", err)
		return
	}
	if *gzip {
		handler = server.GzipHandler(handler, config.GzipMediaTypes)
	}
	if *accessLog {
		handler = server.AccessLogHandler(handler)
	}
	handler = server.RequestIdToCtxHandler(handler)
	fileserver := &http.Server{
		Addr:    ":" + strconv.Itoa(*fileServerPort),
		Handler: handler,
	}
	log.Info().Msgf("Starting fileserver, time elapsed since app start: %s", time.Since(startTime).String())
	errChan <- fileserver.ListenAndServe()
}

func startHealthServer(errChan chan<- error) {
	if *health {
		var handler http.Handler = server.HealthCheckHandler()
		if *healthAccessLog {
			handler = server.AccessLogHandler(handler)
		}
		healthserver := &http.Server{
			Addr:    ":" + strconv.Itoa(*healthPort),
			Handler: handler,
		}
		log.Info().Msgf("Starting healtcheck-server, time elapsed since app start: %s", time.Since(startTime).String())
		errChan <- healthserver.ListenAndServe()
	}
}

func main() {
	httpHeaderConfig, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading http-header-config: See httpHeaderConfig.go for the expected structure.")
	}
	errChan := make(chan error)
	go startFileServer(httpHeaderConfig, errChan)
	go startHealthServer(errChan)
	for err := range errChan {
		log.Fatal().Err(err).Msg("Error starting server: %v")
	}
}
