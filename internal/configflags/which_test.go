package configflags_test

import (
	"flag"
	"reflect"
	"testing"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configflags"
)

func TestWhich(t *testing.T) {
	for _, tc := range []struct {
		label string
		args  []string
		exp   []string
	}{
		{
			label: "return all config flags set",
			args: []string{
				"-method", "POST",
				"-url", "https://benchttp.app?cool=yes",
				"-concurrency", "2",
				"-requests", "3",
				"-requestTimeout", "1s",
				"-globalTimeout", "4s",
			},
			exp: []string{
				"concurrency", "globalTimeout", "method",
				"requestTimeout", "requests", "url",
			},
		},
		{
			label: "do not return config flags not set",
			args:  []string{"-requests", "3"},
			exp:   []string{"requests"},
		},
	} {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)

		configflags.Set(flagset, &config.Global{})

		if err := flagset.Parse(tc.args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		if got := configflags.Which(flagset); !reflect.DeepEqual(got, tc.exp) {
			t.Errorf("\nexp %v\ngot %v", tc.exp, got)
		}
	}
}
