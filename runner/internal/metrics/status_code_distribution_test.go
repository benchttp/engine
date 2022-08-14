package metrics_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/benchttp/engine/runner/internal/metrics"
	"github.com/benchttp/engine/runner/internal/recorder"
)

var validRecords = []recorder.Record{
	{
		Code: 200,
	},
	{
		Code: 200,
	},
	{
		Code: 400,
	},
	{
		Code: 200,
	},
	{
		Code: 400,
	},
	{
		Code: 200,
	},
	{
		Code: 500,
	},
	{
		Code: 200,
	},
	{
		Code: 200,
	},
	{
		Code: 200,
	},
}

func TestComputeStatusCodeDistribution(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := map[string]int{
			"Status1xx": 0,
			"Status2xx": 7,
			"Status3xx": 0,
			"Status4xx": 2,
			"Status5xx": 1,
		}

		got, errs := metrics.ComputeStatusCodesDistribution(validRecords)
		if errs != nil {
			t.Fatalf("want nil error, got %v", errs)
		}

		if reflect.ValueOf(got).IsZero() {
			t.Error("want stats output to be non-zero value, got zero value")
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("StatusCodesDistribution: want %v, got %v", want, got)
		}
	},
	)
	t.Run("invalid status code", func(t *testing.T) {
		invalidRecords := []recorder.Record{
			{
				Code: -1938,
			},
		}
		want := "-1938 is not a valid HTTP status code"

		_, errs := metrics.ComputeStatusCodesDistribution(invalidRecords)
		if errs == nil {
			fmt.Println(errs)
			t.Fatalf("want error, got nil")
		}

		if errs[0].Error() != want {
			t.Errorf("did not get expected error: want %v, got %v", want, errs)
		}
	},
	)
}
