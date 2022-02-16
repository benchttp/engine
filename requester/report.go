package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Report represents the collected results of a benchmark test.
type Report struct {
	Records []Record `json:"records"`
	Length  int      `json:"length"`
	Success int      `json:"success"`
	Fail    int      `json:"fail"`
}

func (rep Report) String() string {
	b, _ := json.MarshalIndent(rep, "", "  ")
	return string(b)
}

// report generates and returns a Report from a previous Run.
func makeReport(records []Record, numErr int) Report {
	return Report{
		Records: records,
		Length:  len(records),
		Success: len(records) - numErr,
		Fail:    numErr,
	}
}

// SendReport sends the report to url. Returns any non-nil error that occurred.;
//
// TODO: move from requester
func SendReport(url string, report Report) error {
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
