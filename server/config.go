package server

import "regexp"

type OperationType string

const SecureRandomIdFileReplacer OperationType = "RANDOM-ID"

type ConfigRaw struct {
	Headers             map[string][]string  `json:"headers"`
	RandomIdReplacerRaw *RandomIdReplacerRaw `json:"random-id-replacer,omitempty"`
	// needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap   map[string]string `json:"media-type-map"`
	GzipMediaTypes []string          `json:"gzip-media-types,omitempty"`
}

type Config struct {
	Headers          map[string][]string
	RandomIdReplacer *RandomIdReplacer
	MediaTypeMap     map[string]string
	GzipMediaTypes   []string
}

type RandomIdReplacerRaw struct {
	FileNamePattern string `json:"file-name-pattern,omitempty"`
	HeaderName      string `json:"header-name,omitempty"`
	VariableName    string `json:"variable-name"`
	MaxReplacements int    `json:"max-replacements"`
}

type RandomIdReplacer struct {
	FileNamePattern regexp.Regexp
	HeaderName      string
	VariableName    string
	MaxReplacements int
}
