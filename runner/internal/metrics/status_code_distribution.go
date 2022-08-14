package metrics

import (
	"github.com/benchttp/engine/runner/internal/recorder"
)

func ComputeStatusCodesDistribution(records []recorder.Record) (statusCodesDistribution map[string]int, errs []error) {
	statusCodesDistribution = map[string]int{
		"Status1xx": 0,
		"Status2xx": 0,
		"Status3xx": 0,
		"Status4xx": 0,
		"Status5xx": 0,
	}

	for _, rec := range records {
		switch rec.Code / 100 {
		case 1:
			statusCodesDistribution["Status1xx"]++
		case 2:
			statusCodesDistribution["Status2xx"]++
		case 3:
			statusCodesDistribution["Status3xx"]++
		case 4:
			statusCodesDistribution["Status4xx"]++
		case 5:
			statusCodesDistribution["Status5xx"]++
		default:
			errs = append(errs, StatusCodeDistributionComputeError(rec.Code))
		}
	}

	return statusCodesDistribution, errs
}
