package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/benchttp/runner/config"
)

// Report represents the collected results of a benchmark test.
type Report struct {
	Config  config.Config `json:"config"`
	Records []Record      `json:"records"`
	Length  int           `json:"length"`
	Success int           `json:"success"`
	Fail    int           `json:"fail"`
}

func (rep Report) String() string {
	b, _ := json.MarshalIndent(rep, "", "  ")
	return string(b)
}

// collect pulls the records from Requester.Records as soon as
// they are available and consumes them to build the report.
// Returns the report when all the records have been collected.
// Requester.collect will blocks until Requester.Records is empty.
func (r *Requester) collect() Report {
	rep := Report{
		Config:  r.config,
		Records: make([]Record, 0, r.config.RunnerOptions.Requests), // Provide capacity if known.
	}

	for rec := range r.records {
		if rec.Error != nil {
			rep.Fail++
		} else {
			rep.Records = append(rep.Records, rec)
		}
	}
	rep.Length = len(rep.Records)
	rep.Success = rep.Length - rep.Fail
	return rep
}

// Report sends the report to url. Returns any non-nil error that occurred.
func (r *Requester) Report(url string, report Report) error {
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(report); err != nil {
		return fmt.Errorf("error sending the report: %s", err)
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("error sending the report: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending the report: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("error sending the report: %s", resp.Status)
	}

	return nil
}

// RunAndReport calls Run and then Report in a single
// invocation. It's useful for simple usecases where the
// caller don't need to known about the Report.
func (r *Requester) RunAndReport(url string) error {
	report := r.Run()

	if err := r.Report(url, report); err != nil {
		return err
	}
	return nil
}
