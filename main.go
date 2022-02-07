package main

import (
	"io/fs"
	"os"

	"github.com/ngergs/webserver/filesystem"
	"github.com/ngergs/webserver/server"
	"github.com/rs/zerolog/log"
)

func main() {
	setup()
	config, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading -config: See server.config.go for the expected structure.")
	}

	var unzipfs fs.ReadFileFS
	var zipfs fs.ReadFileFS
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		memoryFs, err := filesystem.NewMemoryFs(targetDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing read-only filesystem.")
		}
		unzipfs = memoryFs
		if *gzip {
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

	errChan := make(chan error)
	go func() {
		webserver := server.Build(*webServerPort,
			server.FileServerHandler(unzipfs, zipfs, *fallbackFilepath, config),
			server.Caching(unzipfs),
			server.Optional(server.CspReplace(config, unzipfs), config.AngularCspReplace != nil),
			server.Optional(server.Gzip(config), *gzip),
			server.Optional(server.SessionId(config), config.AngularCspReplace != nil),
			server.Header(config),
			server.ValidateClean(),
			server.Optional(server.AccessLog(), *accessLog),
			server.RequestID(),
			server.Timer())
		log.Info().Msgf("Starting webserver server on port %d", *webServerPort)
		errChan <- webserver.ListenAndServe()
	}()
	if *health {
		go func() {
			healthServer := server.Build(*healthPort,
				server.HealthCheckHandler(),
				server.Optional(server.AccessLog(), *healthAccessLog),
			)
			log.Info().Msgf("Starting webserver server on port %d", *healthPort)
			errChan <- healthServer.ListenAndServe()
		}()
	}
	for err := range errChan {
		log.Fatal().Err(err).Msg("Error starting server: %v")
	}
}
