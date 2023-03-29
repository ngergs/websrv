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

var version = "snapshot"
var targetDir string

// readConfig reads the configuration. Order is (least one takes precedence) defaults > config file > env vars.
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

// readConfig reads the configuration. Order is (least one takes precedence) defaults > config file > env vars.
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
