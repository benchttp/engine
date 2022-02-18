package requester

import (
	"crypto/tls"
	"net/http/httptrace"
	"testing"
)

func TestTracer(t *testing.T) {
	t.Run("append events on trace hooks", func(t *testing.T) {
		tracer := newTracer()
		trace := tracer.trace()

		trace.GetConn("")
		trace.DNSDone(httptrace.DNSDoneInfo{})
		trace.ConnectDone("", "", nil)
		trace.TLSHandshakeDone(tls.ConnectionState{}, nil)
		trace.WroteHeaders()
		trace.WroteRequest(httptrace.WroteRequestInfo{})
		trace.GotFirstResponseByte()
		trace.PutIdleConn(nil)

		expEventNames := []string{
			"DNSDone", "ConnectDone", "TLSHandshakeDone", "WroteHeaders",
			"WroteRequest", "GotFirstResponseByte", "PutIdleConn",
		}
		gotEvents := tracer.events

		if len(gotEvents) != len(expEventNames) {
			t.Errorf("missing request events:\nexp %v\n got %v", expEventNames, gotEvents)
		}

		for i, gotEvent := range gotEvents {
			// check event names
			if gotName, expName := gotEvent.Name, expEventNames[i]; gotName != expName {
				t.Errorf("unexpected appended event: exp %s, got %s", expName, gotName)
			}

			// check incremental timestamps
			if i == 0 {
				continue
			}
			if prev := gotEvents[i-1]; gotEvent.Time < prev.Time {
				t.Error("unexpected event time, should be incremental")
			}
		}

		if first, last := gotEvents[0], gotEvents[len(gotEvents)-1]; first.Time == last.Time {
			t.Error("unexpected event times, should be incremental")
		}

		t.Log(tracer.events)
	})
}
