package requester

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/semimpl"
)

// Requester executes the benchmark. It wraps http.Client.
type Requester struct {
	records chan Record // Records provides read access to the results of Requester.Run.

	config config.Config
	client http.Client
}

// New returns a Requester configured with specified Options.
func New(cfg config.Config) *Requester {
	return &Requester{
		records: make(chan Record, cfg.RunnerOptions.Requests),
		config:  cfg,
		client: http.Client{
			// Timeout includes connection time, any redirects, and reading the response body.
			// We may want exclude reading the response body in our benchmark tool.
			Timeout: cfg.Request.Timeout,
		},
	}
}

// Run starts the benchmark test and pipelines the results inside a Report.
// Returns the Report when the test ended and all results have been collected.
func (r *Requester) Run() Report {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.RunnerOptions.GlobalTimeout)

	go func() {
		defer cancel()
		defer close(r.records)
		semimpl.Do(ctx,
			r.config.RunnerOptions.Concurrency,
			r.config.RunnerOptions.Requests,
			r.record,
		)
	}()

	return r.collect()
}

// Record is the summary of a HTTP response. If Record.Error is non-nil,
// the HTTP call failed anywhere from making the request to decoding the
// response body, invalidating the entire response, as it is not a remote
// server error.
type Record struct {
	Time  time.Duration `json:"time"`
	Code  int           `json:"code"`
	Bytes int           `json:"bytes"`
	Error error         `json:"error"`
}

func (r *Requester) record() {
	req, err := r.config.HTTPRequest()
	if err != nil {
		r.records <- Record{Error: err}
		return
	}

	sent := time.Now()

	resp, err := r.client.Do(req)
	if err != nil {
		r.records <- Record{Error: err}
		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		r.records <- Record{Error: err}
		return
	}

	r.records <- Record{
		Code:  resp.StatusCode,
		Time:  time.Since(sent),
		Bytes: len(body),
	}
}
