package requester

import (
	"encoding/json"
	"time"
)

// Report represents the collected results of a benchmark test.
type Report struct {
	Records  []Record      `json:"records"`
	Length   int           `json:"length"`
	Success  int           `json:"success"`
	Fail     int           `json:"fail"`
	Duration time.Duration `json:"duration"`
}

// String returns an indented JSON representation of the report.
func (rep Report) String() string {
	b, _ := json.MarshalIndent(rep, "", "  ")
	return string(b)
}

// Stats returns basic stats about the report's records:
// min duration, max duration, and mean duration.
// It does not replace the remote computing and should only be used
// when a local reporting is needed.
func (rep Report) Stats() (min, max, mean time.Duration) {
	var sum time.Duration
	for _, rec := range rep.Records {
		d := rec.Time
		if d < min || min == 0 {
			min = d
		}
		if d > max {
			max = d
		}
		sum += rec.Time
	}
	return min, max, sum / time.Duration(rep.Length)
}

// newReport generates and returns a Report given a Run dataset.
func newReport(records []Record, numErr int, d time.Duration) Report {
	return Report{
		Records:  records,
		Length:   len(records),
		Success:  len(records) - numErr,
		Fail:     numErr,
		Duration: d,
	}
}
