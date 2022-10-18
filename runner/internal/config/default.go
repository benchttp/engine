package config

import (
	"fmt"
	"net/http"
	"time"
)

var defaultConfig = Global{
	Request: defaultRequest(),
	Runner: Runner{
		Concurrency:    10,
		Requests:       100,
		Interval:       0 * time.Second,
		RequestTimeout: 5 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
}

// Default returns a default config that is safe to use.
func Default() Global {
	return defaultConfig
}

func defaultRequest() *http.Request {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		panic(fmt.Sprintf("benchttp/runner: %s", err))
	}
	return req
}
