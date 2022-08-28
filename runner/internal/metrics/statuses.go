package metrics

import (
	"strconv"

	"github.com/benchttp/engine/runner/internal/recorder"
)

func computeStatusCodesDistribution(records []recorder.Record) map[string]int {
	statuses := map[string]int{}
	for _, rec := range records {
		statuses[strconv.Itoa(rec.Code)]++
	}
	return statuses
}
