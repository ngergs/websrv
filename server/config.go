package server

import (
	"github.com/ngergs/webserver/utils"
)

// Config holds the advanced server config options
type Config struct {
	// Static headers to be set
	Headers           map[string]string        `json:"headers,omitempty"`
	AngularCspReplace *AngularCspReplaceConfig `json:"angular-csp-replace,omitempty"`
	// Mapping of file extension to Media-Type, needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap map[string]string `json:"media-type-map,omitempty"`
	// Media-Types for which gzipping should be applied (if activated and client has set the Accept-Encoding: gzip HTTP-Header)
	GzipMediaTypes []string `json:"gzip-media-types,omitempty"`
}

// AngularCspReplaceConfig holds the config options for fixing the syle-src CSP issue in Angular.
type AngularCspReplaceConfig struct {
	// (secret) placeholder which will be replaced with the session id when serving
	VariableName string `json:"variable-name"`
	// Regex for which files the Variable-Name should be replaced
	FileNamePattern string `json:"file-name-regex,omitempty"`
	// Name of the session-id cookie
	CookieName string `json:"cookie-name"`
	// Max-Age setting for the session-id cookie, 30 seconds should be sufficient
	CookieMaxAge int `json:"cookie-max-age"`
}

// GzipFileExtensions computes the file extensions relevant for gzipping.
func (config *Config) GzipFileExtensions() []string {
	result := make([]string, 0)
	for fileExtension, mediaType := range config.MediaTypeMap {
		if utils.Contains(config.GzipMediaTypes, mediaType) {
			result = append(result, fileExtension)
		}
	}
	return result
}
