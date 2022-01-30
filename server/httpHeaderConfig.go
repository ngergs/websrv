package server

type HttpHeaderConfig struct {
	Headers map[string][]string `json:"headers"`
}
