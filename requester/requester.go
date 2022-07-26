package requester

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/benchttp/engine/internal/dispatcher"
)

const (
	defaultRecordsCap = 1000
)

// Config is the requester config that determines its behavior.
type Config struct {
	Requests       int
	Concurrency    int
	Interval       time.Duration
	RequestTimeout time.Duration
	GlobalTimeout  time.Duration
	OnStateUpdate  func(State)
}

// Requester executes the benchmark. It wraps http.Client.
// It must be initialized with New.
type Requester struct {
	records []Record
	numErr  int
	runErr  error
	start   time.Time
	done    bool

	config       Config
	newTransport func() http.RoundTripper

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

	return &Requester{
		records: make([]Record, 0, recordsCap),
		config:  cfg,
		newTransport: func() http.RoundTripper {
			return newTracer()
		},
	}
}

// Run starts the benchmark test and pipelines the results inside a Report.
// Returns the Report when the test ended and all results have been collected.
func (r *Requester) Run(ctx context.Context, req *http.Request) (Benchmark, error) {
	if err := r.ping(req); err != nil {
		return Benchmark{}, fmt.Errorf("%w: %s", ErrConnection, err)
	}

	var (
		errRun error

		numWorker = r.config.Concurrency
		maxIter   = r.config.Requests
		timeout   = r.config.GlobalTimeout
		interval  = r.config.Interval
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.start = time.Now()
	go r.tickState()

	err := dispatcher.New(numWorker).Do(ctx, maxIter, r.record(req, interval))
	runDuration := time.Since(r.start)

	switch err {
	case nil, context.DeadlineExceeded:
		r.end(err)
	case context.Canceled:
		r.end(err)
		errRun = ErrCanceled
	default:
		return Benchmark{}, err
	}

	return newReport(r.records, r.numErr, runDuration), errRun
}

func (r *Requester) ping(req *http.Request) error {
	client := newClient(r.newTransport(), r.config.RequestTimeout)
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	client.CloseIdleConnections()
	return err
}

// Record is the summary of a HTTP response. If Record.Error is not
// empty string, the HTTP call failed somewhere between sending the request
// to decoding the response body. In that cas invalidating the entire response,
// as it is not a remote server error.
type Record struct {
	Time   time.Duration `json:"time"`
	Code   int           `json:"code"`
	Bytes  int           `json:"bytes"`
	Error  string        `json:"error,omitempty"`
	Events []Event       `json:"events"`
}

func (r *Requester) record(req *http.Request, interval time.Duration) func() {
	return func() {
		// We need new client and request instances each call to this function
		// to make it safe for concurrent use.
		client := newClient(r.newTransport(), r.config.RequestTimeout)
		newReq := cloneRequest(req)

		// Send request
		resp, err := client.Do(newReq)
		if err != nil {
			r.appendRecord(Record{Error: recordErr(err)})
			return
		}

		// Read and close response body
		body, err := readClose(resp)
		if err != nil {
			r.appendRecord(Record{Error: recordErr(err)})
			return
		}

		// Retrieve tracer events and append BodyRead event
		events := []Event{}
		if reqtracer, ok := client.Transport.(*tracer); ok {
			reqtracer.addEventBodyRead()
			events = reqtracer.events
		}

		r.appendRecord(Record{
			Code:   resp.StatusCode,
			Time:   eventsTotalTime(events),
			Bytes:  len(body),
			Events: events,
		})

		r.updateState()
		time.Sleep(interval)
	}
}

func (r *Requester) appendRecord(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	if rec.Error != "" {
		r.numErr++
	}
}

// tickState refreshes the state every second.
func (r *Requester) tickState() {
	ticker := time.NewTicker(time.Second)
	tick := ticker.C
	for {
		if r.done {
			ticker.Stop()
			break
		}
		r.updateState()
		<-tick
	}
}

// updateState calls r.OnStateUpdate with a new computed State
func (r *Requester) updateState() {
	r.safeOnStateUpdate()(r.State())
}

func (r *Requester) safeOnStateUpdate() func(State) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	onStateUpdate := r.config.OnStateUpdate
	if onStateUpdate == nil {
		return func(State) {}
	}
	return onStateUpdate
}

func (r *Requester) end(runErr error) {
	r.mu.Lock()
	r.runErr = runErr
	r.done = true
	r.mu.Unlock()
	r.updateState()
}
