package requester

import (
	"encoding/json"
	"time"
)

// Benchmark represents the collected results of a benchmark test.
type Benchmark struct {
	Records  []Record      `json:"records"`
	Length   int           `json:"length"`
	Success  int           `json:"success"`
	Fail     int           `json:"fail"`
	Duration time.Duration `json:"duration"`
}

// String returns an indented JSON representation of the Benchmark.
func (bk Benchmark) String() string {
	b, _ := json.MarshalIndent(bk, "", "  ")
	return string(b)
}

// Stats returns basic stats about the Benchmark's records:
// min duration, max duration, and mean duration.
// It does not replace the remote computing and should only be used
// when a local reporting is needed.
func (bk Benchmark) Stats() (min, max, mean time.Duration) {
	n := len(bk.Records)
	if n == 0 {
		return 0, 0, 0
	}

	var sum time.Duration
	for _, rec := range bk.Records {
		d := rec.Time
		if d < min || min == 0 {
			min = d
		}
		if d > max {
			max = d
		}
		sum += rec.Time
	}
	return min, max, sum / time.Duration(n)
}

// newReport generates and returns a Benchmark given a Run dataset.
func newReport(records []Record, numErr int, d time.Duration) Benchmark {
	return Benchmark{
		Records:  records,
		Length:   len(records),
		Success:  len(records) - numErr,
		Fail:     numErr,
		Duration: d,
	}
}
