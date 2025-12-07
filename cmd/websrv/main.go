package main

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/ngergs/websrv/v4/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ngergs/websrv/v4/filesystem"
	"github.com/ngergs/websrv/v4/server"
	"github.com/rs/zerolog/log"

	_ "github.com/KimMachineGun/automemlimit"
	_ "go.uber.org/automaxprocs"
)

func main() {
	ll := landlock.V5.BestEffort()
	conf, err := readConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading configuration: See https://github.com/ngergs/websrv/config.yaml for the expected structure.")
	}

	targetDir, err := setup(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Error during initialization")
	}
	if err := landlockFsReadonlyDirs(ll, targetDir, filepath.Join("/", "proc", strconv.Itoa(os.Getpid()), "task")); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	var wg sync.WaitGroup
	sigtermCtx := server.SigTermCtx(context.Background(), time.Duration(conf.ShutdownDelay)*time.Second)
	unzipfs, zipfs := initFs(targetDir, conf)

	errChan := make(chan error)
	var promRegistration *server.PrometheusRegistration
	if conf.Metrics.Enabled {
		promRegistration, err = server.AccessMetricsRegister(prometheus.DefaultRegisterer, conf.Metrics.Namespace)
		if err != nil {
			log.Error().Err(err).Msg("Could not register custom prometheus metrics.")
		}
	}

	r := chi.NewRouter()
	var rateLimitHandler func(http.Handler) http.Handler
	if conf.RateLimit.Enabled {
		if conf.RateLimit.ByIp {
			log.Info().Msgf("Rate limiting per Ip with %d requests per %v", conf.RateLimit.MaxRequests, conf.RateLimit.TimeWindow)
			rateLimitHandler = httprate.LimitByRealIP(conf.RateLimit.MaxRequests, conf.RateLimit.TimeWindow)
		} else {
			log.Info().Msgf("Rate limiting globally with %d requests per %v", conf.RateLimit.MaxRequests, conf.RateLimit.TimeWindow)
			rateLimitHandler = httprate.Limit(conf.RateLimit.MaxRequests, conf.RateLimit.TimeWindow)
		}
	}
	r.Use(
		server.Optional(rateLimitHandler, conf.RateLimit.Enabled),
		server.Optional(server.H2C(conf.Port.H2c), conf.H2C),
		middleware.RequestID,
		middleware.RealIP,
		middleware.Timeout(time.Duration(conf.Timeout.Write)*time.Second),
		server.Optional(server.AccessLog(), conf.Log.AccessLog.General),
		server.Optional(server.AccessMetrics(promRegistration), conf.Metrics.Enabled),
		server.Validate(),
		server.Header(conf.Headers),
		server.Optional(server.SessionId(conf.AngularCspReplace.SessionCookie.Name, time.Duration(conf.AngularCspReplace.SessionCookie.MaxAge)*time.Second),
			conf.AngularCspReplace.Enabled),
		server.Optional(server.CspHeaderReplace(conf.AngularCspReplace.VariableName), conf.AngularCspReplace.Enabled),
		server.Optional(server.Fallback(conf.FallbackPath, http.StatusNotFound), conf.FallbackPath != ""),
	)

	unzipHandler := http.FileServer(http.FS(unzipfs))
	staticZipHandler := server.Caching()(http.FileServer(http.FS(zipfs)))
	dynamicZipHandler := server.Caching()(middleware.Compress(gzip.DefaultCompression, conf.Gzip.MediaTypes...)(unzipHandler))
	var cspPathRegex *regexp.Regexp
	var cspHandler http.Handler
	if conf.AngularCspReplace.Enabled {
		cspPathRegex = regexp.MustCompile(conf.AngularCspReplace.FilePathRegex)
		cspHandler = middleware.Compress(gzip.DefaultCompression, conf.Gzip.MediaTypes...)(
			server.CspFileReplace(conf.AngularCspReplace.VariableName, conf.MediaTypeMap)(unzipHandler))
	}
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cspPathRegex != nil && cspPathRegex.MatchString(r.URL.Path) {
			cspHandler.ServeHTTP(w, r)
			return
		}
		if conf.MemoryFs && conf.Gzip.Enabled {
			if r.URL.Path == conf.FallbackPath {
				w.Header().Set("Content-Encoding", "gzip")
				staticZipHandler.ServeHTTP(w, r)
				return
			}
			mediaType, ok := conf.MediaTypeMap[path.Ext(r.URL.Path)]
			if i := strings.Index(mediaType, ";"); i >= 0 {
				mediaType = mediaType[0:i]
			}
			if r.URL.Path == conf.FallbackPath || (ok && utils.Contains(conf.Gzip.MediaTypes, mediaType)) {
				w.Header().Set("Content-Encoding", "gzip")
				staticZipHandler.ServeHTTP(w, r)
				return
			}
		}
		// gzip not active also will cause the gzipMediaTypes list to be empty so safe to call the generalized handler here
		dynamicZipHandler.ServeHTTP(w, r)
	}))

	webserver := server.Build(conf.Port.Webserver, time.Duration(conf.Timeout.Read)*time.Second,
		time.Duration(conf.Timeout.Write)*time.Second, time.Duration(conf.Timeout.Idle)*time.Second, r)
	log.Info().Msgf("Starting webserver server on port %d", conf.Port.Webserver)
	srvCtx := context.WithValue(sigtermCtx, server.ServerName, "file server")
	server.AddGracefulShutdown(srvCtx, &wg, webserver, time.Duration(conf.Timeout.Shutdown)*time.Second)
	webserver.ListenGoServe(sigtermCtx, errChan)

	if conf.Metrics.Enabled {
		metricsServer := server.Build(conf.Port.Metrics, time.Duration(conf.Timeout.Read)*time.Second,
			time.Duration(conf.Timeout.Write)*time.Second, time.Duration(conf.Timeout.Idle)*time.Second,
			promhttp.Handler(), server.Optional(server.AccessLog(), conf.Log.AccessLog.Metrics))
		metricsCtx := context.WithValue(sigtermCtx, server.ServerName, "prometheus metrics server")
		server.AddGracefulShutdown(metricsCtx, &wg, metricsServer, time.Duration(conf.Timeout.Shutdown)*time.Second)
		metricsServer.ListenGoServe(sigtermCtx, errChan)
		log.Info().Msgf("Listening for prometheus metric scrapes under container port tcp/%s", metricsServer.Addr[1:])
	}

	go logErrors(errChan)

	// stop health server after everything else has stopped
	if conf.Health {
		healthServer := server.Build(conf.Port.Health, time.Duration(conf.Timeout.Read)*time.Second,
			time.Duration(conf.Timeout.Write)*time.Second, time.Duration(conf.Timeout.Idle)*time.Second,
			server.HealthCheckHandler(),
			server.Optional(server.AccessLog(), conf.Log.AccessLog.Health),
		)
		log.Info().Msgf("Starting healthcheck server on port %d", conf.Port.Health)
		healthCtx := context.WithValue(context.Background(), server.ServerName, "health server")
		healthServer.ListenGoServe(sigtermCtx, errChan)
		if err := landlockNetwork(ll); err != nil {
			log.Fatal().Err(err).Msg("")
		}
		// 1 second is sufficient for health checks to shut down
		errChan <- server.ShutdownAfterWaitGroup(healthCtx, &wg, healthServer.Server, time.Duration(1)*time.Second)
	} else {
		if err := landlockNetwork(ll); err != nil {
			log.Fatal().Err(err).Msg("")
		}
		wg.Wait()
	}
}

// initFs loads the non-zipped and zipped fs according to the config
// zipFs is nil if memoryFs or gzipActive are not set
func initFs(targetDir string, conf *config) (unzipfs fs.ReadFileFS, zipfs fs.ReadFileFS) {
	if conf.MemoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		memoryFs, err := filesystem.NewMemoryFs(targetDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing read-only filesystem.")
		}
		unzipfs = memoryFs
		if conf.Gzip.Enabled {
			log.Debug().Msg("Zipping in memory filesystem")
			zipfs, err = memoryFs.Zip()
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

// landlockReadonlyDirs restricts file system access to only readonly permissions for the specified directories
func landlockFsReadonlyDirs(ll landlock.Config, target ...string) error {
	if err := ll.RestrictPaths(landlock.RODirs(target...)); err != nil {
		return fmt.Errorf("error during landlock filesystem restriction: %w", err)
	}
	return nil
}

// landlockNetwork allows no additional tcp connections (connect and bind TCP)
func landlockNetwork(ll landlock.Config) error {
	if err := ll.RestrictNet(); err != nil {
		return fmt.Errorf("error during landlock network restriction: %w", err)
	}
	return nil
}
