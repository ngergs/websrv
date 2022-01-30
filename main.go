package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/ngergs/webserver/v2/memoryfs"
	"github.com/ngergs/webserver/v2/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var accessLog = flag.Bool("access-log", false, "Prints an acess log for the file server endpoint.")
var debugLogging = flag.Bool("debug", false, "Log debug level")
var fallbackFilepath = flag.String("fallback-file", "index.html", "Filepath relative to targetDir which serves as fallback. Set to empty to disable.")
var fileServerPort = flag.Int("port", 8080, "Port under which the fileserver runs.")
var health = flag.Bool("health", true, "Whether to start the health check endpoint (/ under a separate port)")
var healthAccessLog = flag.Bool("health-access-log", false, "Prints an access log for the health check endpoint to stdout.")
var healthPort = flag.Int("health-port", 8081, "Different port under which the health check endpoint runs.")
var help = flag.Bool("help", false, "Prints the help.")
var httpHeaderFile = flag.String("http-header-file", "", "Optional file that contains a set of HTTP headers to be served for each request.")
var memoryFs = flag.Bool("in-memory-fs", false, "Whether to use a in-memory-filesystem. I.e. prefetch the target directory into the heap.")
var prettyLogging = flag.Bool("pretty", false, "Activates zerolog pretty logging")
var targetDir string

func usage() {
	fmt.Printf("Usage: fileserver {options} [target-path]\nOptions:\n")
	flag.PrintDefaults()
}

func init() {
	flag.Parse()
	if *help {
		usage()
		os.Exit(0)
	}
	if *debugLogging {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	if *prettyLogging {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	args := flag.Args()
	if len(args) != 1 {
		log.Error().Msgf("Unexpected number of arguments: %d\n", len(args))
		usage()
		os.Exit(1)
	}
	targetDir = args[0]
}

func startFileServer(httpHeaderConfig *server.HttpHeaderConfig, errChan chan<- error) {
	flag.Parse()
	var filesystem fs.FS
	var err error
	if *memoryFs {
		log.Info().Msg("Using the in-memory-filesystem")
		filesystem, err = memoryfs.New(targetDir)
		if err != nil {
			errChan <- fmt.Errorf("error initializing in-memory-fs: %w", err)
			return
		}
	} else {
		log.Info().Msg("Using nano git os filesystem")
		filesystem = os.DirFS(targetDir)
	}
	var handler http.Handler
	handler, err = server.New(filesystem, *fallbackFilepath, httpHeaderConfig)
	if err != nil {
		errChan <- fmt.Errorf("error initializing webserver handler: %w", err)
		return
	}
	if *accessLog {
		handler = &server.AccessLogWrapper{
			Next: handler,
		}
	}
	fileserver := &http.Server{
		Addr:    ":" + strconv.Itoa(*fileServerPort),
		Handler: handler,
	}
	log.Info().Msg("Starting fileserver")
	errChan <- fileserver.ListenAndServe()
}

func startHealthServer(errChan chan<- error) {
	if *health {
		var handler http.Handler = &server.HealthCheckHandler{}
		if *healthAccessLog {
			handler = &server.AccessLogWrapper{
				Next: handler,
			}
		}
		healthserver := &http.Server{
			Addr:    ":" + strconv.Itoa(*healthPort),
			Handler: handler,
		}
		log.Info().Msg("Starting healtcheck-server")
		errChan <- healthserver.ListenAndServe()
	}
}

func readHttpHeaderConfig() (*server.HttpHeaderConfig, error) {
	if *httpHeaderFile == "" {
		return nil, nil
	}
	file, err := os.Open(*httpHeaderFile)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var result server.HttpHeaderConfig
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func main() {
	httpHeaderConfig, err := readHttpHeaderConfig()
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
