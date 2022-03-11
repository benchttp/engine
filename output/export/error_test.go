package export_test

import (
	"errors"
	"testing"

	"github.com/benchttp/runner/output/export"
)

func TestHTTPResponseError_Is(t *testing.T) {
	for _, tc := range []struct {
		label  string
		target error
		exp    bool
	}{
		{
			label:  "is not any error",
			target: errors.New("any error"),
			exp:    false,
		},
		{
			label:  "is ErrHTTPResponse",
			target: export.ErrHTTPResponse,
			exp:    true,
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			var errHTTPResponse error = &export.HTTPResponseError{Code: 400}
			if got := errors.Is(errHTTPResponse, tc.target); got != tc.exp {
				t.Errorf("exp %v, got %v", tc.exp, got)
			}
		})
	}
}

func TestHTTPResponseError_Error(t *testing.T) {
	t.Run("display accurate error message", func(t *testing.T) {
		err := (&export.HTTPResponseError{Code: 400})
		exp := "export: HTTP response error: status code: 400"
		if got := err.Error(); got != exp {
			t.Errorf("exp\n%s\ngot %s", exp, got)
		}
	})
}
