package recorder

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

// A Config determines the behavior of a Requester.
type Config struct {
	// Requests is the number of requests to send.
	Requests int
	// Concurrency is the maximum number of concurrent requests.
	Concurrency int
	// Interval is the minimum duration between two non-concurrent requests.
	Interval time.Duration
	// RequestTimeout is the timeout for each request sent.
	RequestTimeout time.Duration
	// GlobalTimeout is the timeout for the whole run.
	GlobalTimeout time.Duration
	// OnStateUpdate is called each time the requester state is updated.
	// The requester state is updated each time a requests is done,
	// and every second concurrently.
	OnStateUpdate func(Progress)
}

// Recorder sends requests and records the results via the method Run.
// It must be initialized with New: it won't work otherwise.
type Recorder struct {
	records []Record
	runErr  error
	start   time.Time
	done    bool

	config       Config
	newTransport func() http.RoundTripper

	mu sync.RWMutex
}

// New returns a Requester initialized with the given Config.
func New(cfg Config) *Recorder {
	recordsCap := cfg.Requests
	if recordsCap < 1 {
		recordsCap = defaultRecordsCap
	}

	return &Recorder{
		records: make([]Record, 0, recordsCap),
		config:  cfg,
		newTransport: func() http.RoundTripper {
			return newTracer()
		},
	}
}

// Record clones and sends req n times, or until ctx is done or the global
// timeout  is reached. It gathers the collected results into a Benchmark.
func (r *Recorder) Record(ctx context.Context, req *http.Request) ([]Record, error) {
	if err := r.ping(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnection, err)
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

	err := dispatcher.
		New(numWorker).
		Do(ctx, maxIter, r.recordSingle(req, interval))

	switch err {
	case nil, context.DeadlineExceeded:
		r.end(err)
	case context.Canceled:
		r.end(err)
		errRun = ErrCanceled
	default:
		return nil, err
	}

	return r.records, errRun
}

func (r *Recorder) ping(req *http.Request) error {
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

func (r *Recorder) recordSingle(req *http.Request, interval time.Duration) func() {
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

func (r *Recorder) appendRecord(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
}

// tickState refreshes the state every second.
func (r *Recorder) tickState() {
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
func (r *Recorder) updateState() {
	r.safeOnStateUpdate()(r.Progress())
}

func (r *Recorder) safeOnStateUpdate() func(Progress) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	onStateUpdate := r.config.OnStateUpdate
	if onStateUpdate == nil {
		return func(Progress) {}
	}
	return onStateUpdate
}

func (r *Recorder) end(runErr error) {
	r.mu.Lock()
	r.runErr = runErr
	r.done = true
	r.mu.Unlock()
	r.updateState()
}
