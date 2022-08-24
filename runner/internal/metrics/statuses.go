package metrics

import (
	"strconv"

	"github.com/benchttp/engine/runner/internal/recorder"
)

func computeStatusCodesDistribution(records []recorder.Record) map[string]int {
	statuses := map[string]int{}
	for _, rec := range records {
		s := strconv.Itoa(rec.Code)
		if _, ok := statuses[s]; ok {
			statuses[s]++
		} else {
			statuses[s] = 1
		}
	}
	return statuses
}
