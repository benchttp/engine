package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/benchttp/engine/runner"
)

func TestServer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		config := map[string]interface{}{
			"request": map[string]interface{}{
				"url": "https://example.com",
			},
			"runner": map[string]interface{}{
				"requests":    10,
				"concurrency": 2,
			},
			// "tests": []map[string]interface{}{
			// 	{
			// 		"name":      "maximum response time",
			// 		"metric":    "MAX",
			// 		"predicate": "LT",
			// 		"value":     "100ms",
			// 	},
			// },
		}

		resp, err := makeRunRequest(config)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}

		report := runner.Report{}
		if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
			t.Fatal(err)
		}
	})
}

func makeRunRequest(cfg map[string]interface{}) (*http.Response, error) {
	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	mime := "application/json"
	body := bytes.NewReader(jsonConfig)

	return http.Post(serverRunEndpoint, mime, body)
}
