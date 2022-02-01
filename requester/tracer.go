package requester

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Event is a stage of an outgoing HTTP request associated with a timestamp.
type Event struct {
	Name string        `json:"name"`
	Time time.Duration `json:"time"`
}

// tracer is a http.RoundTripper to be used as a http.Transport
// that records the events of an outgoing HTTP request.
type tracer struct {
	start  time.Time
	events []Event
}

// RoundTrip implements http.RoundTripper. It attaches the client trace
// to the request context and calls http.DefaultTransport.RoundTrip
// with the new created request.
func (p *tracer) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := httptrace.WithClientTrace(r.Context(), p.trace())
	return http.DefaultTransport.RoundTrip(r.WithContext(ctx))
}

// trace returns a http.ClientTrace that timestamps and records the events
// of an outgoing HTTP request.
func (p *tracer) trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(string) {
			p.start = time.Now()
			p.addEvent("GetConn")
		},
		DNSStart: func(httptrace.DNSStartInfo) {
			p.addEvent("DNSStart")
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			p.addEvent("DNSDone")
		},
		ConnectStart: func(string, string) {
			p.addEvent("ConnectStart")
		},
		ConnectDone: func(string, string, error) {
			p.addEvent("ConnectDone")
		},
		GotConn: func(httptrace.GotConnInfo) {
			p.addEvent("GotConn")
		},
		TLSHandshakeStart: func() {
			p.addEvent("TLSHandshakeStart")
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			p.addEvent("TLSHandshakeDone")
		},

		WroteHeaders: func() {
			p.addEvent("WroteHeaders")
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			p.addEvent("WroteRequest")
		},
		GotFirstResponseByte: func() {
			p.addEvent("GotFirstResponseByte")
		},
		PutIdleConn: func(error) {
			p.addEvent("PutIdleConn")
		},
	}
}

// addEvent timestamps and appends and event to the tracer's events slice.
func (p *tracer) addEvent(name string) {
	p.events = append(p.events, Event{Name: name, Time: time.Since(p.start)})
}

// newTracer returns an initialized tracer.
func newTracer() *tracer {
	p := &tracer{events: make([]Event, 0, 20)}
	return p
}
