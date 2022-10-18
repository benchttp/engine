package runner

import (
	"fmt"
	"net/http"
	"time"
)

var defaultConfig = Config{
	Request: defaultRequest(),
	Runner: RunnerConfig{
		Concurrency:    10,
		Requests:       100,
		Interval:       0 * time.Second,
		RequestTimeout: 5 * time.Second,
		GlobalTimeout:  30 * time.Second,
	},
}

// DefaultConfig returns a default config that is safe to use.
func DefaultConfig() Config {
	return defaultConfig
}

func defaultRequest() *http.Request {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		panic(fmt.Sprintf("benchttp/runner: %s", err))
	}
	return req
}
