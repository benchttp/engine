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
	// OnProgress is called each time the requester Progress is updated.
	// The requester Progress is updated each time a request is done,
	// and every second concurrently.
	OnProgress func(Progress)
}

// Recorder sends requests and records the results via the method Run.
// It must be initialized with New: it won't work otherwise.
type Recorder struct {
	records    []Record
	runErr     error
	start      time.Time
	done       bool
	onProgress func(Progress)

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

	onProgress := cfg.OnProgress
	if onProgress == nil {
		onProgress = func(Progress) {}
	}

	return &Recorder{
		records:    make([]Record, 0, recordsCap),
		config:     cfg,
		onProgress: onProgress,
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
	go r.tickProgress()

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

// RelativeTimeEvents returns a copy of the Record.Events
// with each Event.Time as a duration relative to the previous
// Event. For example, if the first Event.Time is 100ms and the
// second Event.Time is 150ms, RelativeTimeEvents evaluates the
// second Event.Time to 50ms.
func (r Record) RelativeTimeEvents() []Event {
	return RelativeTimeEvents(r.Events).Get()
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
			r.mu.Lock()
			reqtracer.addEventBodyRead()
			r.mu.Unlock()
			events = reqtracer.events
		}

		r.appendRecord(Record{
			Code:   resp.StatusCode,
			Time:   eventsTotalTime(events),
			Bytes:  len(body),
			Events: events,
		})

		time.Sleep(interval)
	}
}

func (r *Recorder) appendRecord(rec Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
}

// tickProgress refreshes the Progress every second.
func (r *Recorder) tickProgress() {
	ticker := time.NewTicker(100 * time.Millisecond)
	tick := ticker.C
	for {
		r.mu.RLock()
		done := r.done
		r.mu.RUnlock()
		if done {
			ticker.Stop()
			break
		}
		r.updateProgress()
		<-tick
	}
}

// updateProgress calls r.onProgress with a new computed Progress
func (r *Recorder) updateProgress() {
	r.onProgress(r.Progress())
}

func (r *Recorder) end(runErr error) {
	r.mu.Lock()
	r.runErr = runErr
	r.done = true
	r.mu.Unlock()
	r.updateProgress()
}
