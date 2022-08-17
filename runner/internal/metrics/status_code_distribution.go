package metrics

import (
	"strconv"

	"github.com/benchttp/engine/runner/internal/recorder"
)

func computeStatusCodesDistribution(records []recorder.Record) (statusCodesDistribution map[string]int) {
	statusCodesDistribution = map[string]int{}
	for _, rec := range records {
		stringCode := strconv.Itoa(rec.Code)
		if _, ok := statusCodesDistribution[stringCode]; ok {
			statusCodesDistribution[stringCode]++
		} else {
			statusCodesDistribution[stringCode] = 1
		}
	}
	return statusCodesDistribution
}
