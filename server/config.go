package server

import (
	"github.com/ngergs/webserver/utils"
)

type OperationType string

const SecureRandomIdFileReplacer OperationType = "RANDOM-ID"

type Config struct {
	Headers           map[string]string  `json:"headers,omitempty"`
	AngularCspReplace *AngularCspReplace `json:"angular-csp-replace,omitempty"`
	// needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap   map[string]string `json:"media-type-map,omitempty"`
	GzipMediaTypes []string          `json:"gzip-media-types,omitempty"`
}

type AngularCspReplace struct {
	FileNamePattern string `json:"file-name-regex,omitempty"`
	VariableName    string `json:"variable-name"`
	CookieName      string `json:"cookie-name"`
	CookieMaxAge    int    `json:"cookie-max-age"`
}

func (config *Config) GzipFileExtensions() []string {
	result := make([]string, 0)
	for fileExtension, mediaType := range config.MediaTypeMap {
		if utils.Contains(config.GzipMediaTypes, mediaType) {
			result = append(result, fileExtension)
		}
	}
	return result
}
