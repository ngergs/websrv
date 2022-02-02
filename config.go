package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ngergs/webserver/v2/server"
	"github.com/ngergs/webserver/v2/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var accessLog = flag.Bool("access-log", false, "Prints an acess log for the file server endpoint.")
var debugLogging = flag.Bool("debug", false, "Log debug level")
var configFile = flag.String("config-file", "", "Optional file that contains more involved config settings, see server/config.go for the structure.")
var fallbackFilepath = flag.String("fallback-file", "index.html", "Filepath relative to targetDir which serves as fallback. Set to empty to disable.")
var webServerPort = flag.Int("port", 8080, "Port under which the webserver runs.")
var gzip = flag.Bool("gzip", true, "Whether to send gzip encoded response. See config-file for setting the detailed types. As default gzip is used when activated for test/css, text/html and application/javascript")
var health = flag.Bool("health", true, "Whether to start the health check endpoint (/ under a separate port)")
var healthAccessLog = flag.Bool("health-access-log", false, "Prints an access log for the health check endpoint to stdout.")
var healthPort = flag.Int("health-port", 8081, "Different port under which the health check endpoint runs.")
var help = flag.Bool("help", false, "Prints the help.")
var memoryFs = flag.Bool("in-memory-fs", false, "Whether to use a in-memory-filesystem. I.e. prefetch the target directory into the heap.")
var prettyLogging = flag.Bool("pretty", false, "Activates zerolog pretty logging")
var targetDir string

var defaultGzipMediaTypes []string = []string{"application/javascript", "text/css", "text/html; charset=UTF-8"}

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

func readConfig() (*server.Config, []string, error) {
	if *configFile == "" {
		return &server.Config{
			GzipMediaTypes: defaultGzipMediaTypes,
		}, nil, nil
	}
	file, err := os.Open(*configFile)
	if err != nil {
		return nil, nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	var config server.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, nil, err
	}
	gzipFileExtensions := []string{}
	if !*gzip {
		config.GzipMediaTypes = []string{}
	} else {
		if config.GzipMediaTypes == nil {
			config.GzipMediaTypes = defaultGzipMediaTypes
		}
		for fileExtension, mediaType := range config.MediaTypeMap {
			if utils.Contains(config.GzipMediaTypes, mediaType) {
				gzipFileExtensions = append(gzipFileExtensions, fileExtension)
			}
		}
	}
	return &config, gzipFileExtensions, nil
}
