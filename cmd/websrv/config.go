package main

// config is the general configuration struct
type config struct {
	// Log configures log properties
	Log logConfig `koanf:"log"`
	// Headers is a map of static HTTP response headers
	Headers map[string]string `koanf:"headers"`
	// MediaTypeMap is a map of file extensions like ".jk" to corresponding media types.
	MediaTypeMap map[string]string `koanf:"mediatypes"`
	// FallbackPath is the path that should be used as an alternative on HTTP 404 responses. Set to empty to disable.
	FallbackPath string `koanf:"fallback"`
	// Metrics holds the configuration for prometheus metrics
	Metrics metricsConfig `koanf:"metrics"`
	// MemoryFs enables the in-memory filesystem
	MemoryFs bool `koanf:"memoryfs"`
	// H2C enables the h2c (unencrypted HTTP2) endpoint
	H2C bool `koanf:"h2c"`
	// Health enables the health endpoint
	Health bool `koanf:"health"`
	// Port holds the configuration for various TCP ports
	Port portConfig `koanf:"port"`
	// Gzip holds the configuration for gzip compression handling
	Gzip gzipConfig `koanf:"gzip"`
	// Timeout holds the configuration for various timeouts
	Timeout timeoutConfig `koanf:"timeout"`
	// ShutdownDelay is the number of seconds to wait before executing a graceful shutdown
	ShutdownDelay int `koanf:"shutdowndelay"`
	// AngularCspReplace holds the configuration for angular csp fix
	AngularCspReplace angularCspReplaceConfig `koanf:"angularcsp"`
}

// logConfig holds configuration regarding logging
type logConfig struct {
	// Level sets the log level. Valid values are debug, info, warn, error
	Level string `koanf:"level"`
	// Pretty enables pretty printed log (non-json)
	Pretty bool `koanf:"pretty"`
	// AccessLog determines whether to print an access log
	AccessLog accessLogConfig `koanf:"access"`
}

// accessLogConfig configures which accesses logs to enable
type accessLogConfig struct {
	// General enables the general access log
	General bool `koanf:"general"`
	// Health enables the health endpoint access log
	Health bool `koanf:"health"`
	// Metrics enables the metrics endpoint access log
	Metrics bool `koanf:"metrics"`
}

// metricsConfig holds the prometheus metrics configuration
type metricsConfig struct {
	// Enabled activates the prometheus metrics endpoint
	Enabled bool `koanf:"enabled"`
	// Namespace is the prometheus namespace
	Namespace string `koanf:"namespace"`
}

// portConfig holds configurations for various TCP ports
type portConfig struct {
	// Webserver is the TCP port for the main web server
	Webserver uint16 `koanf:"webserver"`
	// Health is the TCP port for the health endpoint
	Health uint16 `koanf:"health"`
	// Metrics is the TCP port for the prometheus metrocs
	Metrics uint16 `koanf:"metrics"`
	// H2c is the TCP port for h2c (unecncrypted http2)
	H2c uint16 `koanf:"h2c"`
}

// gzipConfig holds configuration for gzip response compression
type gzipConfig struct {
	// Enabled activates the gzip response compression
	Enabled bool `koanf:"enabled"`
	// CompressionLevel is the amount of compression, values are between 1 and 9
	CompressionLevel int `koanf:"compression"`
	// MediaTypes is a slice of media type (according to the response HTTP Content-Type header) that should be compressed
	MediaTypes []string `koanf:"mediatypes"`
}

// timeoutConfig holds various timeouts
type timeoutConfig struct {
	// Idle is the idle timeout in seconds
	Idle int `koanf:"idle"`
	// Read is the request read timeout in seconds
	Read int `koanf:"read"`
	// Write is the request response timeout in seconds
	Write int `koanf:"write"`
	// Shutdown is the graceful shutdown timeout in seconds
	Shutdown int `koanf:"shutdown"`
}

// angularCspReplaceConfig holds the configuration for the angular csp replace fix
type angularCspReplaceConfig struct {
	// Enabled activates the angular csp fix
	Enabled bool `koanf:"enabled"`
	// FilePathRegex is a regular expression for the files whether the VariableName should be replace, like "^/main.*\.js$"
	FilePathRegex string `koanf:"filepath"`
	// VariableName is the string that should be replaced with the Session-Id value
	VariableName string `koanf:"variable"`
	// Cookie holds the config for the session cookie
	SessionCookie cookieConfig `koanf:"sessioncookie"`
}

// angularCspReplaceConfig holds the configuration for the angular csp replace fix
type cookieConfig struct {
	// Name is the name of the cookie that will hold the Session-ID
	Name string `koanf:"name"`
	// MaxAge is the max age of the Session-ID cookie
	MaxAge int `koanf:"maxage"`
}

//nolint:mnd
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
