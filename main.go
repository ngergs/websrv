package main

import (
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ngergs/webserver/v2/filesystems"
	"github.com/ngergs/webserver/v2/server"
	"github.com/rs/zerolog/log"
)

func getFilesystem(config *server.Config) (fs.FS, error) {
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		filesystem, err := filesystems.NewMemoryFs(targetDir)
		if err != nil {
			return nil, err
		}
		return filesystem, nil
	}
	log.Info().Msg("Using the os filesystem")
	return os.DirFS(targetDir), nil
}

func startFileServer(config *server.Config, filesystem fs.FS, errChan chan<- error, handlerSetups ...HandlerSetup) {
	var handler http.Handler
	handler = server.New(filesystem, *fallbackFilepath, config)
	for _, handlerSetup := range handlerSetups {
		handler = handlerSetup(handler)
	}
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
	config, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading -config: See server.config.go for the expected structure.")
	}
	filesystem, err := getFilesystem(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error preparing read-only filesystem.")
	}
	errChan := make(chan error)
	go startFileServer(config, filesystem, errChan,
		Caching(filesystem),
		FileReplace(config, filesystem),
		Header(config),
		Gzip(config, *gzip),
		ValidateClean(),
		AccessLog(*accessLog),
		RequestID(),
	)
	go startHealthServer(errChan)
	for err := range errChan {
		log.Fatal().Err(err).Msg("Error starting server: %v")
	}
}
