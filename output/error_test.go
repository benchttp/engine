package output_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/benchttp/runner/output"
	"github.com/benchttp/runner/output/export"
)

func TestExportError_HasAuthError(t *testing.T) {
	for _, tc := range []struct {
		label string
		errs  []error
		exp   bool
	}{
		{
			label: "return false with no errors",
			errs:  []error{},
			exp:   false,
		},
		{
			label: "return false without auth errors",
			errs:  []error{errors.New("any error")},
			exp:   false,
		},
		{
			label: "return true with ErrNoToken",
			errs:  []error{output.ErrNoToken},
			exp:   true,
		},
		{
			label: "return true with Unauthorized error",
			errs:  []error{export.ErrHTTPResponse.WithCode(http.StatusUnauthorized)},
			exp:   true,
		},
		{
			label: "return true with mixed errors including auth error",
			errs: []error{
				errors.New("any error"),
				export.ErrHTTPResponse.WithCode(http.StatusUnauthorized),
				errors.New("any error"),
			},
			exp: true,
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			errExport := &output.ExportError{Errors: tc.errs}
			if got := errExport.HasAuthError(); got != tc.exp {
				t.Errorf("exp %v, got %v", tc.exp, got)
			}
		})
	}
}
