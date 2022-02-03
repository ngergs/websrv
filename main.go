package main

import (
	"os"

	"github.com/ngergs/webserver/v2/filesystem"
	"github.com/ngergs/webserver/v2/server"
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
			server.CspReplace(config, fs),
			server.Header(config),
			server.Optional(server.Gzip(config), *gzip),
			server.ValidateClean(),
			server.Optional(server.AccessLog(), *accessLog),
			server.SessionId(config, 10),
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
