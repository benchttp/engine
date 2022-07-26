# Output structure

```go
type Run struct {
	MetricsAggregate struct {
		Min, Max, Mean, StdDev    time.Duration
		Deciles                   map[int]float64
		NumRequestFailures        int
		StatusCodeDistribution    map[string]int
		RequestEventsDistribution map[requester.Event]int
	}

	Tests struct {
		GlobalPass bool
		Results    []struct {
			Name   string
			Pass   bool
			Reason string
		}
	}

	Metadata struct {
		Config     config.Global
		FinishedAt time.Time
		StartedAt  time.Time
		RunEvents  []struct {
			Name string // or "Type"
			Time time.Time
		}
	}
}
```
