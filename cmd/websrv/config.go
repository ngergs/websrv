package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	stdlog "log"

	"github.com/ngergs/websrv/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var version = "snapshot"
var accessLog = flag.Bool("access-log", false, "Prints an acess log for the file server endpoint.")
var debugLogging = flag.Bool("debug", false, "Log debug level")
var configFile = flag.String("config-file", "", "Optional file that contains more involved config settings, see server/config.go for the structure.")
var fallbackFilepath = flag.String("fallback-file", "index.html", "Filepath relative to targetDir which serves as fallback. Set to empty to disable.")
var webServerPort = flag.Int("port", 8080, "Port under which the webserver runs.")
var gzipActive = flag.Bool("gzip", true, "Whether to send gzip encoded response. See config-file for setting the detailed types. As default gzip is used when activated for test/css, text/html and application/javascript")
var gzipCompressionLevel = flag.Int("gzip-level", gzip.DefaultCompression, "The compression level used for gzip compression. See the golang gzip documentation for details. Only applies to on-the-fly compression. The in-memory-fs (when used) uses for static files always gzip.BestCompression")
var health = flag.Bool("health", true, "Whether to start the health check endpoint (/ under a separate port)")
var healthAccessLog = flag.Bool("health-access-log", false, "Prints an access log for the health check endpoint to stdout.")
var healthPort = flag.Int("health-port", 8081, "Different port under which the health check endpoint runs.")
var help = flag.Bool("help", false, "Prints the help.")
var memoryFs = flag.Bool("in-memory-fs", false, "Whether to use a in-memory-filesystem. I.e. prefetch the target directory into the heap.")
var prettyLogging = flag.Bool("pretty", false, "Activates zerolog pretty logging")
var idleTimeout = flag.Int("idle-timeout", 30, "Timeout for idle TCP connections with keep-alive in seconds.")
var readTimeout = flag.Int("read-timeout", 10, "Timeout to read the entire request in seconds.")
var shutdownTimeout = flag.Int("shutdown-timeout", 10, "Timeout for the graceful shutdown in seconds.")
var shutdownDelay = flag.Int("shutdown-delay", 5, "Delay before shutting down the server in seconds. To make sure that the load balancing of the surrounding infrastructure had time to update.")
var writeTimeout = flag.Int("write-timeout", 10, "Timeout to write the complete response in seconds.")
var targetDir string

var defaultGzipMediaTypes = []string{"application/javascript", "text/css", "text/html; charset=UTF-8"}
var defaultMediaTypeMap = map[string]string{
	".js":    "application/javascript",
	".css":   "text/css",
	".html":  "text/html; charset=UTF-8",
	".jpg":   "image/jpeg",
	".avif":  "image/avif",
	".jxl":   "image/jxl",
	".ttf":   "font/ttf",
	".woff2": "font/woff2",
	".txt":   "text/plain",
}

func setup() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s {options} [target-path]\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *help {
		flag.Usage()
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
		flag.Usage()
		os.Exit(1)
	}
	targetDir = args[0]

	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
	log.Info().Msgf("This is websrv version %s", version)
}

// GetDefaultConfig returns the default configuration used when none is provided.
func GetDefaultConfig() *server.Config {
	return &server.Config{
		GzipMediaTypes: defaultGzipMediaTypes,
		MediaTypeMap:   defaultMediaTypeMap,
	}
}

// readConfig reads and deserializes the configFile flag parameter.
// Returns a default Configuration with default mediatype file extension mappings as well as default gzip media types. if the configFile flag parameter has not been set.
func readConfig() (*server.Config, error) {
	if *configFile == "" {
		return GetDefaultConfig(), nil
	}
	file, err := os.Open(*configFile)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var config server.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.MediaTypeMap == nil {
		config.MediaTypeMap = defaultMediaTypeMap
	}

	if !*gzipActive {
		config.GzipMediaTypes = []string{}
		return &config, nil
	}
	if config.GzipMediaTypes == nil {
		config.GzipMediaTypes = defaultGzipMediaTypes
	}
	return &config, nil
}
