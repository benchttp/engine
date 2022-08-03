package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/benchttp/engine/runner"
	"github.com/gorilla/websocket"
)

func TestServer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		const numRequest = 4
		const failingMaxDuration = 10 * time.Millisecond

		cfg := map[string]interface{}{
			"request": map[string]interface{}{
				"url": "https://example.com",
			},
			"runner": map[string]interface{}{
				"requests":    numRequest,
				"concurrency": 2,
			},
			"tests": []map[string]interface{}{
				{
					"name":      "some passing test",
					"field":     "TOTAL_COUNT",
					"predicate": "EQ",
					"target":    numRequest,
				},
				{
					"name":      "some failing test",
					"field":     "MAX",
					"predicate": "LT",
					"target":    failingMaxDuration.String(),
				},
			},
		}

		ws, err := connectWS()
		if err != nil {
			t.Fatal(err)
		}

		report, err := runAndWait(ws, cfg)
		if err != nil {
			t.Fatal(err)
		}

		// some non-exhaustive healthchecks on retrieved Report

		t.Run("requests count", func(t *testing.T) {
			if report.Metrics.TotalCount != numRequest {
				t.Errorf("got bad report: %+v", report)
			}
		})

		t.Run("test suite results", func(t *testing.T) {
			if gotlen, explen := len(report.Tests.Results), 2; gotlen != explen {
				t.Errorf("len(report.Tests.Results): exp %d, got %d", explen, gotlen)
			}
			if gotPass, expPass := report.Tests.Pass, false; gotPass != expPass {
				t.Errorf("report.Tests.Pass: exp %v, got %v", expPass, gotPass)
			}
			if gotPass, expPass := report.Tests.Results[0].Pass, true; gotPass != expPass {
				t.Errorf("report.Tests.Results[0].Pass: exp %v, got %v", expPass, gotPass)
			}
			if gotPass, expPass := report.Tests.Results[1].Pass, false; gotPass != expPass {
				t.Errorf("report.Tests.Results[1].Pass: exp %v, got %v", expPass, gotPass)
			}
		})

		if t.Failed() {
			t.Logf("\n%+v", report)
		}
	})
}

func connectWS() (*websocket.Conn, error) {
	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     "run",
		RawQuery: url.Values{"access_token": []string{serverDummyToken}}.Encode(),
	}

	ws, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ws, nil
}

func runAndWait(
	ws *websocket.Conn,
	cfg map[string]interface{},
) (runner.Report, error) {
	if err := ws.WriteJSON(map[string]interface{}{
		"action": "run",
		"data":   cfg,
	}); err != nil {
		return runner.Report{}, err
	}

	for {
		var msg map[string]interface{}
		if err := ws.ReadJSON(&msg); err != nil {
			return runner.Report{}, err
		}

		switch msg["state"] {
		case "error":
			if errStr, exists := msg["error"]; exists {
				return runner.Report{}, fmt.Errorf("got error message: %s", errStr)
			}
			return runner.Report{}, fmt.Errorf("got unexpected error message: %+v", msg)

		case "progress":
			// TODO: tests on progress event
			continue

		case "done":
			if errStr, exists := msg["error"]; exists {
				return runner.Report{}, fmt.Errorf("event done: got error message: %s", errStr)
			}

			data, hasData := msg["data"]
			if !hasData {
				return runner.Report{}, errors.New("event done: no data")
			}

			dataMap, validType := data.(map[string]interface{})
			if !validType {
				return runner.Report{}, fmt.Errorf("event done: bad data type: %+v", dataMap)
			}

			return mapToReport(dataMap)
		}
	}
}

func mapToReport(m map[string]interface{}) (runner.Report, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return runner.Report{}, fmt.Errorf("mapToReport: %w", err)
	}

	report := runner.Report{}
	if err := json.Unmarshal(b, &report); err != nil {
		return runner.Report{}, fmt.Errorf("mapToReport: %w", err)
	}

	return report, nil
}
