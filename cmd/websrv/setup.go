package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-viper/mapstructure/v2"
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

var (
	ErrInvalidLogLevel        = errors.New("invalid loglevel, only error, warn, info and debug are valid")
	ErrInvalidNumberArguments = errors.New("invalid number of argument, has to be 1")

	version = "snapshot"
)

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
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, envPrefix)), "_", ".")
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("error loading config from env vars:%w", err)
	}

	// fail on config settings that do not match any internal setting, see https://github.com/knadh/koanf/issues/189
	err = k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{DecoderConfig: &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc()),
		Metadata:         nil,
		Result:           &conf,
		ErrorUnused:      true,
		WeaklyTypedInput: true,
	}})
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling collected config: %w", err)
	}
	return &conf, nil
}

// setup uses the configuration to set log levels, it also reads input args and returns the targetDir
func setup(conf *config) (string, error) {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s {options} [target-path]\nOptions:\n", os.Args[0])
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
		return "", fmt.Errorf("%w: %s", ErrInvalidLogLevel, conf.Log.Level)
	}
	if conf.Log.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
	log.Info().Msgf("This is websrv version %s", version)

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return "", fmt.Errorf("%w: %d\n", ErrInvalidNumberArguments, len(args))
	}

	return args[0], nil
}
