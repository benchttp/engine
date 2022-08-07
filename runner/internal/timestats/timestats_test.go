package timestats_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/engine/runner/internal/recorder"
	"github.com/benchttp/engine/runner/internal/timestats"
)

var validRecords = []recorder.Record{
	{
		Time: time.Duration(100.000000),
	},
	{
		Time: time.Duration(200.000000),
	},
	{
		Time: time.Duration(300.000000),
	},
	{
		Time: time.Duration(400.000000),
	},
	{
		Time: time.Duration(200.000000),
	},
	{
		Time: time.Duration(100.000000),
	},
	{
		Time: time.Duration(200.000000),
	},
	{
		Time: time.Duration(300.000000),
	},
	{
		Time: time.Duration(400.000000),
	},
	{
		Time: time.Duration(200.000000),
	},
}

func TestCompute(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := timestats.TimeStats{
			Min:    100,
			Max:    400,
			Avg:    240,
			Median: 200,
			StdDev: 101,
		}

		got, errs := timestats.Compute(validRecords)

		if errs != nil {
			t.Fatalf("want nil error, got %v", errs)
		}

		for _, stat := range []struct {
			name string
			want time.Duration
			got  time.Duration
		}{
			{"min", want.Min, got.Min},
			{"max", want.Max, got.Max},
			{"avg", want.Avg, got.Avg},
		} {
			if !approxEqualTime(stat.got, stat.want, 1) {
				t.Errorf("%s: want %d, got %d", stat.name, stat.want, stat.got)
			}
		}
	})

	t.Run("passing invalid dataset returns error", func(t *testing.T) {
		for _, testcase := range []struct {
			name string
			data []recorder.Record
			want error
			zero bool
		}{
			{
				name: "empty dataset",
				data: []recorder.Record{},
				want: timestats.ErrEmptySlice,
				zero: true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				res, errs := timestats.Compute(testcase.data)

				if errs == nil {
					t.Error("want error, got none")
				}

				if !containsError(errs, testcase.want) {
					t.Errorf("want %T, got %+v", testcase.want, errs)
				}

				switch {
				case testcase.zero && !reflect.ValueOf(res).IsZero():
					t.Errorf("want stats output to be zero value, got %+v", res)
				case !testcase.zero && reflect.ValueOf(res).IsZero():
					t.Error("want stats output to be non-zero value, got zero value")
				}
			})
		}
	})
}

// approxEqual returns true if val is equal to target with a margin of error.
func approxEqualTime(val, target, margin time.Duration) bool {
	return val >= target-margin && val <= target+margin
}

// contains checks if an error is present in a slice of errors
func containsError(errs []error, err error) bool {
	for _, v := range errs {
		if v.Error() == err.Error() {
			return true
		}
	}
	return false
}
