package metrics

import (
	"reflect"
	"testing"

	"github.com/benchttp/engine/runner/internal/recorder"
)

var validRecords = []recorder.Record{
	{Code: 200},
	{Code: 200},
	{Code: 400},
	{Code: 200},
	{Code: 400},
	{Code: 200},
	{Code: 500},
	{Code: 200},
	{Code: 200},
	{Code: 200},
}

func TestComputeStatusCodeDistribution(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := map[string]int{"200": 7, "400": 2, "500": 1}

		got := computeStatusCodesDistribution(validRecords)

		if reflect.ValueOf(got).IsZero() {
			t.Error("want stats output to be non-zero value, got zero value")
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("StatusCodesDistribution: want %v, got %v", want, got)
		}
	},
	)
}
