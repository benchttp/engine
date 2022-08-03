package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/benchttp/engine/runner"
	"github.com/gorilla/websocket"
)

func TestServer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		const numRequest = 4

		cfg := map[string]interface{}{
			"request": map[string]interface{}{
				"url": "https://example.com",
			},
			"runner": map[string]interface{}{
				"requests":    numRequest,
				"concurrency": 2,
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

		gotRequestCount := report.Metrics.TotalCount
		if gotRequestCount != numRequest {
			t.Errorf("got bad report: %+v", report)
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

		if msg["state"] == "progress" {
			// TODO: tests on progress event
			continue
		}

		if msg["state"] == "done" {
			errStr, isErr := msg["error"]
			if isErr {
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
