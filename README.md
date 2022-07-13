# websrv

This is a little webserver implementation using only dependencies from the go standard library (besides testing and logging).
The webserver is supposed to serve a folder containing e.g. a static website. 

The server package contains a collection of http.Handler implementations which may be reused in other projects. 
The filesystem package contains a readonly in-memory-filesystem implementation.

## Server package features
Logs are (without -pretty option) are provided in a GCP compatible JSON format.

The following limited handler features are provided in the server package:
* FileServer: implements the main logic for providing the requested file or an optional fallback-file.
* Headers: Static Headers can be easily configured.
* Gzip: Supported for on-demand and in-memory-filesystem.
* Caching: Support via ETag and If-None-Match HTTP-Headers
* Request-ID: All logs provide a unique request-id to correlate different logs to a given request.
* Access-Log: Basic access-logging.
* Timer: Detailed timing reports, set log level to debug to get the output.
* CspReplace: See [my blog](https://ngergs.de/content/angular/style-csp-fix) about fixing Angular CSP regarding style-src.

## Usage

## Docker container
You can use the ngergs/websrv docker container, the entrypoint is set to the websrv executable.

## Binary build
All releases starting with v1.4.3 have binary builds attached to them for linux, windows and osx.

### Compilation from Source
Compile from source:
```bash
git clone https://github.com/ngergs/websrv
go build
```

## Usage
The path to this folder has to be provided as command line argument. There are a number of various optional settings.

```
Usage: ./websrv {options} [target-path]
Options:
  -access-log
        Prints an acess log for the file server endpoint.
  -config-file string
        Optional file that contains more involved config settings, see server/config.go for the structure.
  -debug
        Log debug level
  -fallback-file string
        Filepath relative to targetDir which serves as fallback. Set to empty to disable. (default "index.html")
  -gzip
        Whether to send gzip encoded response. See config-file for setting the detailed types. As default gzip is used when activated for test/css, text/html and application/javascript (default true)
  -gzip-level int
        The compression level used for gzip compression. See the golang gzip documentation for details. Only applies to on-the-fly compression. The in-memory-fs (when used) uses for static files always gzip.BestCompression (default -1)
  -health
        Whether to start the health check endpoint (/ under a separate port) (default true)
  -health-access-log
        Prints an access log for the health check endpoint to stdout.
  -health-port int
        Different port under which the health check endpoint runs. (default 8081)
  -help
        Prints the help.
  -idle-timeout int
        Timeout for idle TCP connections with keep-alive in seconds. (default 30)
  -in-memory-fs
        Whether to use a in-memory-filesystem. I.e. prefetch the target directory into the heap.
  -port int
        Port under which the webserver runs. (default 8080)
  -pretty
        Activates zerolog pretty logging
  -read-timeout int
        Timeout to read the entire request in seconds. (default 10)
  -shutdown-delay int
        Delay before shutting down the server in seconds. To make sure that the load balancing of the surrounding infrastructure had time to update. (default 5)
  -shutdown-timeout int
        Timeout for the graceful shutdown in seconds. (default 10)
  -write-timeout int
        Timeout to write the complete response in seconds. (default 10)
```

## Advanced configs
More involved configs can be provided via the config-file option. See [config-example.json](config-example.json) for an exmaple. The structure is as follows:

```go
type Config struct {
	// Static headers to be set
	Headers           map[string]string  `json:"headers,omitempty"`
	AngularCspReplace *AngularCspReplace `json:"angular-csp-replace,omitempty"`
	// Mapping of file extension to Media-Type, needed due to https://github.com/golang/go/issues/32350
	MediaTypeMap map[string]string `json:"media-type-map,omitempty"`
	// Media-Types for which gzipping should be applied (if activated and client has set the Accept-Encoding: gzip HTTP-Header)
	GzipMediaTypes []string `json:"gzip-media-types,omitempty"`
}

type AngularCspReplace struct {
	// (secret) placeholder which will be replaced with the session id when serving
	VariableName    string `json:"variable-name"`
	// Regex for which files the Variable-Name should be replaced
	FileNamePattern string `json:"file-name-regex,omitempty"`
	// Name of the session-id cookie
	CookieName      string `json:"cookie-name"`
	// Max-Age setting for the session-id cookie, 30 seconds should be sufficient
	CookieMaxAge    int    `json:"cookie-max-age"`
}
```
