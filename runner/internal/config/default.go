package config

import (
	"net/http"
	"net/url"
	"time"
)

var defaultConfig = Global{
	Request: Request{
		Method: "GET",
		URL:    &url.URL{},
		Header: http.Header{},
		Body:   RequestBody{},
	},
	Runner: Runner{
		Concurrency:    10,
		Requests:       100,
		Interval:       0 * time.Second,
		RequestTimeout: 5 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
	Output: Output{
		Silent: false,
	},
}

// Default returns a default config that is safe to use.
func Default() Global {
	return defaultConfig
}
