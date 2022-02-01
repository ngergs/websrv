package server

import (
	"net/http"
	"strings"
)

type HeaderHandler struct {
	Next    http.Handler
	Headers map[string]string
	Replace *FromHeaderReplace
}

func (handler *HeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// set static headers
	if handler.Headers != nil {
		for k, v := range handler.Headers {
			w.Header().Set(k, v)
		}
	}

	// now replace parts if required
	if handler.Replace != nil {
		newValue := getLastHeaderEntry(r, handler.Replace.SourceHeaderName)
		header := w.Header().Get(handler.Replace.TargetHeaderName)
		header = strings.Replace(header, handler.Replace.VariableName, newValue, -1)
		w.Header().Set(handler.Replace.TargetHeaderName, header)
	}
	handler.Next.ServeHTTP(w, r)
}

func getLastHeaderEntry(r *http.Request, headerName string) string {
	values := r.Header.Values(headerName)
	if len(values) > 0 {
		return values[len(values)-1] //select last value as nginx just appends
	}
	return ""
}
