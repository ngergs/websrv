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

func startServer(name string, port int, handler http.Handler, errChan chan<- error, handlerSetups ...HandlerMiddleware) {
	//	handler = server.New(filesystem, *fallbackFilepath, config)
	for _, handlerSetup := range handlerSetups {
		handler = handlerSetup(handler)
	}
	fileserver := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: handler,
	}
	log.Info().Msgf("Starting %s server on port %d, time elapsed since app start: %s", name, port, time.Since(startTime).String())
	errChan <- fileserver.ListenAndServe()
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
	webserver := server.New(filesystem, *fallbackFilepath, config)
	go startServer("webserver", *webServerPort, webserver, errChan,
		Caching(filesystem),
		FileReplace(config, filesystem),
		Header(config),
		Gzip(config, *gzip),
		ValidateClean(),
		AccessLog(*accessLog),
		RequestID(),
	)
	if *health {
		go startServer("healthserver", *healthPort, server.HealthCheckHandler(), errChan,
			AccessLog(*healthAccessLog),
		)
	}
	for err := range errChan {
		log.Fatal().Err(err).Msg("Error starting server: %v")
	}
}
