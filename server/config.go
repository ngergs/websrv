package server

type OperationType string

const SecureRandomIdFileReplacer OperationType = "RANDOM-ID"

type Config struct {
	Headers           map[string]string  `json:"headers"`
	FromHeaderReplace *FromHeaderReplace `json:"from-header-replacer,omitempty"`
	// needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap   map[string]string `json:"media-type-map"`
	GzipMediaTypes []string          `json:"gzip-media-types,omitempty"`
}

type FromHeaderReplace struct {
	FileNamePattern  string `json:"file-name-pattern,omitempty"`
	SourceHeaderName string `json:"source-header-name,omitempty"`
	TargetHeaderName string `json:"target-header-name,omitempty"`
	VariableName     string `json:"variable-name"`
}
