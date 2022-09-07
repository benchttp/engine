package metrics_test

import (
	"testing"

	"github.com/benchttp/engine/runner/internal/metrics"
)

func TestField_Type(t *testing.T) {
	cases := []struct {
		name    string
		fieldID string
		exp     string
	}{
		{
			name:    "nested struct",
			fieldID: "ResponseTimes.Mean",
			exp:     "time.Duration",
		},
		{
			name:    "method call",
			fieldID: "RequestSuccessCount",
			exp:     "int",
		},
		{
			name:    "nil map",
			fieldID: "RequestEventTimes.ConnectDone.Mean",
			exp:     "time.Duration",
		},
		{
			name:    "nil slice",
			fieldID: "Records.0.ResponseTime",
			exp:     "time.Duration",
		},
	}

	for _, c := range cases {
		if got := metrics.Field(c.fieldID).Type(); got != c.exp {
			t.Errorf("bad type: exp %s, got %s", c.exp, got)
		}
	}
}
