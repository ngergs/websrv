# websrv

This is a little webserver implementation bsed on [chi](https://github.com/go-chi/chi).
The webserver is supposed to serve a folder containing e.g. a static website and is suited to serve a SPA.

The server package contains a collection of http.Handler implementations which may be reused in other projects. 
The filesystem package contains a readonly in-memory-filesystem implementation.

## Server package features
Logs are (without -pretty option) are provided in a GCP compatible JSON format.

The following middleware handler features are provided in the server package:
* Fallback: Handler that falls back on a configured default path when retrieving a specified set of status codes from the next handler. 
Very useful for serving a SPA.
* Headers: Static Headers can be easily configured.
* Caching: Support via ETag and If-None-Match HTTP-Headers
* Access-Log: Basic access-logging formatted in a [GCP-compatible](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry) way.
* CspReplace and SessionCookie: See [my blog](https://ngergs.de/content/angular/style-csp-fix) about fixing Angular CSP regarding style-src.

## Usage

## Docker container
You can use the ngergs/websrv docker container, the entrypoint is set to the websrv executable.

## Binary build
All releases starting with v1.4.3 have binary builds attached to them for linux, windows and osx.

### Compilation from Source
Compile from source:
```bash
git clone https://github.com/ngergs/websrv
go build ./cmd/websrv
```

## Usage
The path to the folder to be served has to be provided as command line argument.
```
Usage: ./websrv {options} [target-path]
Options:
  -conf string
        config file to load
```

## Config file settings 
There are a number of various optional settings configured via config files.
The config options and documentation can be found in the [config.yaml](config.yaml). There is also an [example configuration](example/config.yaml).

## Config from env
All config settings can be also set via environment variables. Environment variables take precedence over config file settings. All env config vars start with `WEBSRV` and follow with
the upper-cased config-setting name. You can use underscores to set nested values. To e.g. set the value of
```yaml
log:
  level: info
```
the following configuration variable can be used: `WEBSRV_LOG_LEVEL=info`.