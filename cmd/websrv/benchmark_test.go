package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/ngergs/websrv/v3/filesystem"
	"github.com/ngergs/websrv/v3/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func BenchmarkServer(b *testing.B) {
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	config := defaultConfig
	config.AngularCspReplace = angularCspReplaceConfig{
		FilePathRegex: ".*",
		VariableName:  "testt",
		SessionCookie: cookieConfig{
			Name:   "Nonce-ID",
			MaxAge: 10,
		},
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
		middleware.Timeout(time.Duration(config.Timeout.Write)*time.Second),
		server.Optional(server.AccessLog(), config.Log.AccessLog.General),
		server.Validate(),
		server.Header(config.Headers),
		server.SessionId(config.AngularCspReplace.SessionCookie.Name, time.Duration(config.AngularCspReplace.SessionCookie.MaxAge)*time.Second),
		server.CspHeaderReplace(config.AngularCspReplace.VariableName),
		server.Fallback("/", http.StatusNotFound),
	)
	unzipHandler := http.FileServer(http.FS(fs))
	staticZipHandler := server.Caching()(http.FileServer(http.FS(zipfs)))
	dynamicZipHandler := server.Caching()(middleware.Compress(5, config.Gzip.MediaTypes...)(unzipHandler))
	cspPathRegex := regexp.MustCompile(config.AngularCspReplace.FilePathRegex)
	cspHandler := server.CspFileReplace(config.AngularCspReplace.VariableName, config.MediaTypeMap)(unzipHandler)
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cspPathRegex.MatchString(r.URL.Path) {
			cspHandler.ServeHTTP(w, r)
		} else {
			if config.Metrics.Enabled && config.Gzip.Enabled {
				w.Header().Set("Content-Encoding", "gzip")
				staticZipHandler.ServeHTTP(w, r)
			} else {
				dynamicZipHandler.ServeHTTP(w, r)
			}
		}
	}))
	webserver := server.Build(config.Port.Webserver, time.Duration(1)*time.Second, time.Duration(1)*time.Second, time.Duration(1)*time.Second, r)
	defer func() {
		err := webserver.Shutdown(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to shutdown server")
		}
	}()
	go func() {
		log.Info().Msgf("Starting webserver server on port %d", config.Port.Webserver)
		errChan <- webserver.ListenAndServe()
	}()
	// give server some time to wake up
	time.Sleep(time.Duration(100) * time.Millisecond)
	client := &http.Client{}
	defer client.CloseIdleConnections()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+webserver.Addr+"/dummy_random.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case err = <-errChan:
			log.Error().Err(err).Msg("Webserver error")
			b.Fail()
			return
		default:
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Msg("Get request failed")
				b.Fail()
				return
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("Failed to close request body")
				b.Fail()
				return
			}
		}
	}

}
