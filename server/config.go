package server

import "regexp"

type OperationType string

const SecureRandomIdFileReplacer OperationType = "RANDOM-ID"

type ConfigRaw struct {
	Headers               map[string]string      `json:"headers"`
	FromHeaderReplacerRaw *FromHeaderReplacerRaw `json:"from-header-replacer,omitempty"`
	// needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap   map[string]string `json:"media-type-map"`
	GzipMediaTypes []string          `json:"gzip-media-types,omitempty"`
}

type Config struct {
	Headers            map[string]string
	FromHeaderReplacer *FromHeaderReplacer
	MediaTypeMap       map[string]string
	GzipMediaTypes     []string
}

type FromHeaderReplacerRaw struct {
	FileNamePattern  string `json:"file-name-pattern,omitempty"`
	SourceHeaderName string `json:"source-header-name,omitempty"`
	TargetHeaderName string `json:"target-header-name,omitempty"`
	VariableName     string `json:"variable-name"`
	MaxReplacements  int    `json:"max-replacements"`
}

type FromHeaderReplacer struct {
	FileNamePattern  regexp.Regexp
	SourceHeaderName string
	TargetHeaderName string
	VariableName     string
	MaxReplacements  int
}
