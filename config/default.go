package config

import (
	"net/url"
	"time"
)

var defaultConfig = Config{
	Request: Request{
		Method:  "GET",
		URL:     &url.URL{},
		Timeout: 10 * time.Second,
	},
	RunnerOptions: RunnerOptions{
		Concurrency:   1,
		Requests:      -1, // Use GlobalTimeout as exit condition.
		GlobalTimeout: 30 * time.Second,
	},
}
