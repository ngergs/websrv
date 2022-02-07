package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ngergs/webserver/filesystem"
	"github.com/ngergs/webserver/server"
	"github.com/rs/zerolog/log"
)

func BenchmarkServer(b *testing.B) {
	config := GetDefaultConfig()
	config.AngularCspReplace = &server.AngularCspReplace{
		FileNamePattern: ".*",
		VariableName:    "testt",
		CookieName:      "Nonce-Id",
		CookieMaxAge:    10,
	}
	fs, err := filesystem.NewMemoryFs("benchmark")
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing in-memory-fs for benchmark")
	}
	zipfs, err := fs.Zip(config.GzipFileExtensions())
	if err != nil {
		log.Fatal().Err(err).Msg("error zipping in-memory-fs for benchmark")
	}
	errChan := make(chan error)
	webserver := server.Build(*webServerPort,
		server.FileServerHandler(fs, zipfs, *fallbackFilepath, config),
		server.Caching(fs),
		server.CspReplace(config, fs),
		server.Gzip(config),
		server.SessionId(config),
		server.Header(config),
		server.ValidateClean(),
		server.AccessLog(),
		server.RequestID(),
		server.Timer())
	defer webserver.Shutdown(context.Background())
	go func() {
		log.Info().Msgf("Starting webserver server on port %d", *webServerPort)
		errChan <- webserver.ListenAndServe()
	}()
	// give server some time to wake up
	time.Sleep(time.Duration(100) * time.Millisecond)
	client := &http.Client{}
	defer client.CloseIdleConnections()
	r, _ := http.NewRequest("GET", "http://"+webserver.Addr+"/dummy_random.js", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	for i := 0; i < b.N; i++ {
		select {
		case err = <-errChan:
			log.Fatal().Err(err).Msg("Webserver error")
		default:
			resp, err := client.Do(r)
			if err != nil {
				log.Fatal().Err(err).Msg("Get request failed")
			}
			resp.Body.Close()
		}
	}

}
