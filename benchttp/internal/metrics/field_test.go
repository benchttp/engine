package metrics_test

import (
	"testing"

	"github.com/benchttp/engine/benchttp/internal/metrics"
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
		t.Run(c.name, func(t *testing.T) {
			if got := metrics.Field(c.fieldID).Type(); got != c.exp {
				t.Errorf("bad type: exp %q, got %q", c.exp, got)
			}
		})
	}
}

func TestField_Validate(t *testing.T) {
	cases := []struct {
		name     string
		fieldID  string
		expError string
	}{
		{
			name:     "valid field",
			fieldID:  "ResponseTimes.Mean",
			expError: "",
		},
		{
			name:     "invalid field",
			fieldID:  "Marcel.Patulacci",
			expError: "metrics: unknown field: Marcel.Patulacci",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			shouldError := c.expError != ""
			err := metrics.Field(c.fieldID).Validate()
			if shouldError {
				if err == nil {
					t.Error("expect non-nil error, got nil")
				}
				if got := err.Error(); got != c.expError {
					t.Errorf("exp %q\ngot %q", c.expError, got)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %q", err.Error())
			}
		})
	}
}
