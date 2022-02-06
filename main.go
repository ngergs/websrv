package main

import (
	"os"

	"github.com/ngergs/webserver/filesystem"
	"github.com/ngergs/webserver/server"
	"github.com/rs/zerolog/log"
)

func main() {
	config, gzipFileExtensions, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading -config: See server.config.go for the expected structure.")
	}

	var fs filesystem.ZipFs
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		fs, err = filesystem.NewMemoryFs(targetDir, gzipFileExtensions)
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing read-only filesystem.")
		}
	} else {
		log.Info().Msg("Using the os filesystem")
		fs = filesystem.FromUnzippedFs(os.DirFS(targetDir))
	}

	errChan := make(chan error)
	go func() {
		errChan <- server.Start("webserver", *webServerPort, errChan,
			server.FileServerHandler(fs, *fallbackFilepath, config, *memoryFs),
			server.Caching(fs),
			server.Optional(server.CspReplace(config, fs), config.AngularCspReplace != nil),
			server.Optional(server.Gzip(config), *gzip),
			server.Optional(server.SessionId(config), config.AngularCspReplace != nil),
			server.Header(config),
			server.ValidateClean(),
			server.Optional(server.AccessLog(), *accessLog),
			server.RequestID(),
			server.Timer(),
		)
	}()
	if *health {
		go func() {
			errChan <- server.Start("healthserver", *healthPort, errChan,
				server.HealthCheckHandler(),
				server.Optional(server.AccessLog(), *healthAccessLog),
			)
		}()
	}
	for err := range errChan {
		log.Fatal().Err(err).Msg("Error starting server: %v")
	}
}
