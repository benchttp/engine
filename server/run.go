package server

import (
	"context"
	"io"
	"net/http"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configparse"
	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/requester"
)

type runHandler struct{}

func (h runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/run" {
		http.NotFound(w, r)
		return
	}

	readBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cfg, err := configparse.JSON(readBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	report, err := doRun(silentConfig(cfg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonReport, err := report.JSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonReport)
}

func doRun(cfg config.Global) (output.Report, error) {
	// Retrieve HTTP request generated by the config
	req, err := cfg.Request.Value()
	if err != nil {
		return output.Report{}, err
	}

	// Run benchmark
	bk, err := requester.New(requesterConfig(cfg)).Run(context.Background(), req)
	if err != nil {
		return output.Report{}, err
	}

	return *output.New(bk, cfg, ""), nil
}

func silentConfig(cfg config.Global) config.Global {
	cfg.Output = config.Output{
		Silent:   true,
		Out:      []config.OutputStrategy{},
		Template: "",
	}
	return cfg
}

// requesterConfig returns a requester.Config generated from cfg.
func requesterConfig(cfg config.Global) requester.Config {
	return requester.Config{
		Requests:       cfg.Runner.Requests,
		Concurrency:    cfg.Runner.Concurrency,
		Interval:       cfg.Runner.Interval,
		RequestTimeout: cfg.Runner.RequestTimeout,
		GlobalTimeout:  cfg.Runner.GlobalTimeout,
		Silent:         cfg.Output.Silent,
	}
}
