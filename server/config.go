package server

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
	Domain          string `json:"domain,omitempty"`
	FileNamePattern string `json:"file-name-pattern,omitempty"`
	VariableName    string `json:"variable-name"`
}
