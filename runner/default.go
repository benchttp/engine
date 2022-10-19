package runner

import (
	"fmt"
	"net/http"
	"time"
)

// DefaultRunner returns a default Runner that is safe to use.
func DefaultRunner() Runner {
	return defaultRunner
}

var defaultRunner = Runner{
	Request: defaultRequest(),

	Concurrency:    10,
	Requests:       100,
	Interval:       0 * time.Second,
	RequestTimeout: 5 * time.Second,
	GlobalTimeout:  30 * time.Second,
}

func defaultRequest() *http.Request {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		panic(fmt.Sprintf("benchttp/runner: %s", err))
	}
	return req
}
