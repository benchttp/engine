package recorder

import (
	"math"
	"time"
)

// Event is a stage of an outgoing HTTP request associated with a timestamp.
type Event struct {
	Name string        `json:"name"`
	Time time.Duration `json:"time"`
}

type RelativeTimeEvents []Event

func (e RelativeTimeEvents) Get() []Event {
	d := make([]Event, len(e))
	for i, event := range e {
		switch i {
		case 0:
			d[i] = event
		default:
			d[i] = event.withRelativeTime(e[i-1])
		}
	}
	return d
}

func (e Event) withRelativeTime(ref Event) Event {
	return Event{
		Name: e.Name,
		Time: diff(e.Time, ref.Time),
	}
}

// diff returns the time.Duration difference between a and b.
// a and b need not to be passed in specific order. The difference
// is expressed in absolute value.
func diff(a, b time.Duration) time.Duration {
	d := a - b
	switch {
	case d >= 0:
		return d
	case d == math.MinInt64:
		return math.MaxInt64
	default:
		return -d
	}
}
