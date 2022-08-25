package recorder_test

import (
	"reflect"
	"testing"

	"github.com/benchttp/engine/runner/internal/recorder"
)

func TestRelativeTimeEvents(t *testing.T) {
	e := recorder.RelativeTimeEvents{
		{Time: 0},
		{Time: 100},
		{Time: 110},
		{Time: 200},
	}

	got := e.Get()

	want := []recorder.Event{{Time: 0}, {Time: 100}, {Time: 10}, {Time: 90}}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("incorrect diff: want %v, got %v", want, got)
	}
}
