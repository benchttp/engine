package requester

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/benchttp/runner/dispatcher"
)

const (
	defaultRecordsCap = 1000
)

type Config struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
}

// Requester executes the benchmark. It wraps http.Client.
type Requester struct {
	records []Record
	numErr  int
	runErr  error
	start   time.Time
	done    bool

	config Config
	client http.Client
	tracer *tracer

	mu sync.RWMutex
}

// New returns a Requester initialized with cfg. cfg is assumed valid:
// it is the caller's responsibility to ensure cfg is valid using
// cfg.Validate.
func New(cfg Config) *Requester {
	recordsCap := cfg.Requests
	if recordsCap < 1 {
		recordsCap = defaultRecordsCap
	}

	tracer := newTracer()

	return &Requester{
		records: make([]Record, 0, recordsCap),
		config:  cfg,
		tracer:  tracer,
		client: http.Client{
			// Timeout includes connection time, any redirects, and reading
			// the response body.
			// We may want exclude reading the response body in our benchmark tool.
			Timeout: cfg.RequestTimeout,

			// tracer keeps track of all events of the current request.
			Transport: tracer,
		},
	}
}

// Run starts the benchmark test and pipelines the results inside a Report.
// Returns the Report when the test ended and all results have been collected.
func (r *Requester) Run(req *http.Request) (Report, error) {
	if err := r.ping(req); err != nil {
		return Report{}, fmt.Errorf("%w: %s", ErrConnection, err)
	}

	var (
		numWorker   = r.config.Concurrency
		maxIter     = r.config.Requests
		timeout     = r.config.GlobalTimeout
		interval    = r.config.Interval
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)

	defer cancel()

	r.start = time.Now()
	go r.refreshState()

	switch err := dispatcher.New(numWorker).Do(ctx, maxIter, r.record(req, interval)); err {
	case nil, context.Canceled, context.DeadlineExceeded:
		r.end(err)
	default:
		return Report{}, err
	}

	return makeReport(r.records, r.numErr), nil
}

func (r *Requester) ping(req *http.Request) error {
	resp, err := r.client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	return err
}

// Record is the summary of a HTTP response. If Record.Error is non-nil,
// the HTTP call failed anywhere from making the request to decoding the
// response body, invalidating the entire response, as it is not a remote
// server error.
type Record struct {
	Time   time.Duration `json:"time"`
	Code   int           `json:"code"`
	Bytes  int           `json:"bytes"`
	Error  error         `json:"error,omitempty"`
	Events []Event       `json:"events"`
}

func (r *Requester) record(req *http.Request, interval time.Duration) func() {
	return func() {
		// It is necessary to clone the request because one request with a non-nil body cannot be used in concurrent threads
		reqClone, err := cloneRequest(req)
		if err != nil {
			r.appendRecord(Record{Error: ErrRequestBody})
			return
		}

		sent := time.Now()

		resp, err := r.client.Do(reqClone)
		if err != nil {
			r.appendRecord(Record{Error: err})
			return
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			r.appendRecord(Record{Error: err})
			return
		}

		duration := time.Since(sent)

		r.appendRecord(Record{
			Code:   resp.StatusCode,
			Time:   duration,
			Bytes:  len(body),
			Events: r.tracer.events,
		})

		fmt.Print(r.state())
		time.Sleep(interval)
	}
}

func (r *Requester) appendRecord(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	if rec.Error != nil {
		r.numErr++
	}
}

func (r *Requester) refreshState() {
	ticker := time.NewTicker(time.Second)
	tick := ticker.C
	for {
		if r.done {
			ticker.Stop()
			break
		}
		r.printState()
		<-tick
	}
}

func (r *Requester) end(runErr error) {
	r.mu.Lock()
	r.runErr = runErr
	r.done = true
	r.mu.Unlock()
	r.printState()
}

func (r *Requester) printState() {
	fmt.Print(r.state())
}

// cloneRequest fully clones a http.Request by also cloning the body via Request.GetBody
func cloneRequest(req *http.Request) (*http.Request, error) {
	reqClone := req.Clone(req.Context())
	if req.Body != nil {
		bodyClone, err := req.GetBody()
		if err != nil {
			return nil, ErrRequestBody
		}
		reqClone.Body = bodyClone
	}
	return reqClone, nil
}
