package server

import (
	"net/url"

	"github.com/benchttp/engine/runner"
)

func config() runner.Config {
	rqurl, _ := url.ParseRequestURI("https://example.com")
	config := runner.DefaultConfig()
	config.Request.URL = rqurl
	return config
}
