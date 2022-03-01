package configflags_test

import (
	"flag"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configflags"
)

func TestSet(t *testing.T) {
	t.Run("default to base config", func(t *testing.T) {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)
		args := []string{} // no args

		cfg := config.Default()
		configflags.Set(flagset, &cfg)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		if exp := config.Default(); !reflect.DeepEqual(cfg, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, cfg)
		}
	})

	t.Run("set config with flags values", func(t *testing.T) {
		flagset := flag.NewFlagSet("run", flag.ExitOnError)
		args := []string{
			"-method", "POST",
			"-url", "https://benchttp.app?cool=yes",
			"-header", "Content-Type:application/json",
			"-body", "raw:hello",
			"-requests", "1",
			"-concurrency", "2",
			"-interval", "3s",
			"-requestTimeout", "4s",
			"-globalTimeout", "5s",
			"-out", "stdout,json",
			"-silent",
			"-template", "{{ .Report.Length }}",
		}

		cfg := config.Global{}
		configflags.Set(flagset, &cfg)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		exp := config.Global{
			Request: config.Request{
				Method: "POST",
				Header: http.Header{"Content-Type": {"application/json"}},
				Body:   config.Body{Type: "raw", Content: []byte("hello")},
			}.WithURL("https://benchttp.app?cool=yes"),
			Runner: config.Runner{
				Requests:       1,
				Concurrency:    2,
				Interval:       3 * time.Second,
				RequestTimeout: 4 * time.Second,
				GlobalTimeout:  5 * time.Second,
			},
			Output: config.Output{
				Out:      []config.OutputStrategy{config.OutputStdout, config.OutputJSON},
				Silent:   true,
				Template: "{{ .Report.Length }}",
			},
		}

		if !reflect.DeepEqual(cfg, exp) {
			t.Errorf("\nexp %#v\ngot %#v", exp, cfg)
		}
	})
}
