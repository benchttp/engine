package request

import (
	"context"
	"net/http"
	"time"

	"github.com/benchttp/runner/record"
	"github.com/benchttp/runner/semimpl"
)

func doRequest(url string, timeout time.Duration) record.Record {
	client := http.Client{
		// Timeout includes connection time, any redirects, and reading the response body.
		// We may want exclude reading the response body in our benchmark tool.
		Timeout: timeout,
	}

	start := time.Now()

	resp, err := client.Get(url) //nolint:bodyclose
	end := time.Since(start)
	if err != nil {
		return record.Record{Error: err}
	}

	return record.New(resp, end)
}

// Do launches a goroutine to ping url as soon as a thread is
// available and collects the results as they come in.
// The value of concurrency limits the number of concurrent threads.
// Once all requests have been made or on done signal from ctx,
// waits for goroutines to end and returns the collected records.
func Do(ctx context.Context, requests, concurrency int, url string, timeout time.Duration) []record.Record {
	records := record.NewSafeSlice(requests)

	semimpl.Do(ctx, concurrency, requests, func() {
		records.Append(doRequest(url, timeout))
	})

	return records.Slice()
}
