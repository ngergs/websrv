package main

import (
	"flag"
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"os"
	"strings"

	stdlog "log"

	"github.com/rs/zerolog/log"
)

type config struct {
	// Whether to print an access log
	AccessLog         accessLogConfig         `koanf:"access-log"`
	Log               logConfig               `koanf:"log"`
	Headers           map[string]string       `koanf:"headers"`
	MediaTypeMap      map[string]string       `koanf:"mediatype-map"`
	FallbackPath      string                  `koanf:"fallback-path"`
	Metrics           metricsConfig           `koanf:"metrics"`
	MemoryFs          bool                    `koanf:"memory-fs"`
	H2C               bool                    `koanf:"h2c"`
	Health            bool                    `koanf:"health"`
	Port              portConfig              `koanf:"port"`
	Gzip              gzipConfig              `koanf:"gzip"`
	Timeout           timeoutConfig           `koanf:"timeout"`
	ShutdownDelay     int                     `koanf:"shutdown-delay"`
	AngularCspReplace angularCspReplaceConfig `koanf:"angular-csp-replace"`
}

type accessLogConfig struct {
	General bool `koanf:"general"`
	Health  bool `koanf:"health"`
	Metrics bool `koanf:"metrics"`
}

type logConfig struct {
	Level  string `koanf:"level""`
	Pretty bool   `koanf:"pretty"`
}

type metricsConfig struct {
	Enabled   bool   `koanf:"enabled"`
	AccessLog bool   `koanf:"accesslog"`
	Namespace string `koanf:"namespace"`
}

type portConfig struct {
	Webserver int `koanf:"webserver"`
	Health    int `koanf:"health"`
	Metrics   int `koanf:"metrics"`
	H2c       int `koanf:"h2c"`
}

type gzipConfig struct {
	Enabled          bool     `koanf:"enabled"`
	CompressionLevel int      `koanf:"compression"`
	MediaTypes       []string `koanf:"mediatypes"`
}

type timeoutConfig struct {
	Idle     int `koanf:"idle"`
	Read     int `koanf:"read"`
	Write    int `koanf:"write"`
	Shutdown int `koanf:"shutdown"`
}

type angularCspReplaceConfig struct {
	Enabled       bool   `koanf:"enabled"`
	FilePathRegex string `koanf:"file-path-regex"`
	VariableName  string `koanf:"variable-name"`
	CookieName    string `koanf:"cookie-name"`
	CookieMaxAge  int    `koanf:"cookie-max-age"`
}

var version = "snapshot"
var targetDir string

var defaultConfig = config{
	Log: logConfig{Level: "info"},
	Port: portConfig{
		Webserver: 8080,
		Health:    8081,
		Metrics:   9090,
		H2c:       443,
	},
	Gzip: gzipConfig{
		CompressionLevel: 5,
		MediaTypes:       []string{"text/css", "text/html", "text/javascript", "font/tff"},
	},
	MediaTypeMap: map[string]string{
		".js":    "application/javascript",
		".css":   "text/css",
		".html":  "text/html; charset=UTF-8",
		".jpg":   "image/jpeg",
		".avif":  "image/avif",
		".jxl":   "image/jxl",
		".ttf":   "font/ttf",
		".woff2": "font/woff2",
		".txt":   "text/plain",
	},
	Metrics:       metricsConfig{Namespace: "websrv"},
	Timeout:       timeoutConfig{Idle: 30, Read: 10, Write: 10, Shutdown: 5},
	ShutdownDelay: 5,
}

// readConfig reads the configuration. Order is (least one takes precedence) defaults > config file > env vars > cli args.
func readConfig() (*config, error) {
	k := koanf.New(".")
	var conf config

	// Load defaults
	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		return nil, fmt.Errorf("error loading  default values (internal error): %w", err)
	}

	// Load config from file
	confFile := flag.String("conf", "", "config file to load")
	flag.Parse()
	if *confFile != "" {
		if err := k.Load(file.Provider(*confFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading config file: %w", err)
		}
	}

	// Load config from env
	envPrefix := "WEBSRV_"
	err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, envPrefix)), "_", ".", -1)
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("error loading config from env vars:%w", err)
	}

	if err := k.Unmarshal("", &conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling collected config: %w", err)
	}
	return &conf, nil
}

// readConfig reads the configuration. Order is (least one takes precedence) defaults > config file > env vars > cli args.
func setup(conf *config) error {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s {options} [target-path]\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	switch conf.Log.Level {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		return fmt.Errorf("invalid loglevel, only error, warn, info and debug are valid. Set value: %s", conf.Log.Level)
	}
	if conf.Log.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return fmt.Errorf("unexpected number of arguments: %d\n", len(args))
	}
	targetDir = args[0]

	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
	log.Info().Msgf("This is websrv version %s", version)
	return nil
}
