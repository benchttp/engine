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

// report generates and returns a Report from a previous Run.
func makeReport(cfg config.Config, records []Record, numErr int) Report {
	return Report{
		Config:  cfg,
		Records: records,
		Length:  len(records),
		Success: len(records) - numErr,
		Fail:    numErr,
	}
}

// SendReport sends the report to url. Returns any non-nil error that occurred.
func (r *Requester) SendReport(url string, report Report) error {
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(report); err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrReporting, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: %s", ErrReporting, resp.Status)
	}

	return nil
}

// RunAndSendReport calls Run and then Report in a single
// invocation. It's useful for simple usecases where the
// caller don't need to known about the Report.
func (r *Requester) RunAndSendReport(url string) error {
	report, err := r.Run()
	if err != nil {
		return err
	}

	if err := r.SendReport(url, report); err != nil {
		return err
	}

	return nil
}
