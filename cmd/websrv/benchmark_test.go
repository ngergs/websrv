package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/ngergs/websrv/v2/filesystem"
	"github.com/ngergs/websrv/v2/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func BenchmarkServer(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	config := GetDefaultConfig()
	config.AngularCspReplace = &server.AngularCspReplaceConfig{
		FilePathPattern: ".*",
		VariableName:    "testt",
		CookieName:      "Nonce-Id",
		CookieMaxAge:    10,
	}
	fs, err := filesystem.NewMemoryFs("../../test/benchmark")
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing in-memory-fs for benchmark")
	}
	zipfs, err := fs.Zip()
	if err != nil {
		log.Fatal().Err(err).Msg("error zipping in-memory-fs for benchmark")
	}
	errChan := make(chan error)
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Timeout(time.Duration(*writeTimeout)*time.Second),
		server.Optional(server.AccessLog(), *accessLog),
		server.Validate(),
		server.Header(config),
		server.SessionId(config),
		server.CspHeaderReplace(config),
		server.Fallback("/", http.StatusNotFound),
	)
	unzipHandler := http.FileServer(http.FS(fs))
	staticZipHandler := server.Caching()(http.FileServer(http.FS(zipfs)))
	dynamicZipHandler := server.Caching()(middleware.Compress(5, config.GzipMediaTypes...)(unzipHandler))
	cspPathRegex := regexp.MustCompile(config.AngularCspReplace.FilePathPattern)
	cspHandler := server.CspFileReplace(config)(unzipHandler)
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cspPathRegex.MatchString(r.URL.Path) {
			cspHandler.ServeHTTP(w, r)
		} else {
			if *memoryFs && *gzipActive {
				w.Header().Set("Content-Encoding", "gzip")
				staticZipHandler.ServeHTTP(w, r)
			} else {
				dynamicZipHandler.ServeHTTP(w, r)
			}
		}
	}))
	webserver := server.Build(*webServerPort, time.Duration(1)*time.Second, time.Duration(1)*time.Second, time.Duration(1)*time.Second, r)
	defer func() {
		err := webserver.Shutdown(context.Background())
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to shutdown server")
		}
	}()
	go func() {
		log.Info().Msgf("Starting webserver server on port %d", *webServerPort)
		errChan <- webserver.ListenAndServe()
	}()
	// give server some time to wake up
	time.Sleep(time.Duration(100) * time.Millisecond)
	client := &http.Client{}
	defer client.CloseIdleConnections()
	req, _ := http.NewRequest("GET", "http://"+webserver.Addr+"/dummy_random.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case err = <-errChan:
			log.Fatal().Err(err).Msg("Webserver error")
		default:
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal().Err(err).Msg("Get request failed")
			}
			err = resp.Body.Close()
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to close request body")
			}
		}
	}

}
