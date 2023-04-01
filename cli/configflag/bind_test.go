package configflag_test

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/benchttp/engine/benchttp"
	"github.com/benchttp/engine/benchttptest"
	"github.com/benchttp/engine/configio"

	"github.com/benchttp/engine/cli/configflag"
)

func TestBind(t *testing.T) {
	t.Run("default to zero runner", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ExitOnError)
		args := []string{} // no args

		b := configio.Builder{}
		configflag.Bind(flagset, &b)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		benchttptest.AssertEqualRunners(t, benchttp.Runner{}, b.Runner())
	})

	t.Run("set config with flags values", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ExitOnError)
		args := []string{
			"-method", "POST",
			"-url", "https://example.com?a=b",
			"-header", "API_KEY:abc",
			"-header", "Accept:text/html",
			"-header", "Accept:application/json",
			"-body", "raw:hello",
			"-requests", "1",
			"-concurrency", "2",
			"-interval", "3s",
			"-requestTimeout", "4s",
			"-globalTimeout", "5s",
		}

		b := configio.Builder{}
		configflag.Bind(flagset, &b)
		if err := flagset.Parse(args); err != nil {
			t.Fatal(err) // critical error, stop the test
		}

		benchttptest.AssertEqualRunners(t,
			benchttp.Runner{
				Request: &http.Request{
					Method: "POST",
					URL:    mustParseURL("https://example.com?a=b"),
					Header: http.Header{
						"API_KEY": []string{"abc"},
						"Accept":  []string{"text/html", "application/json"},
					},
					Body: io.NopCloser(bytes.NewBufferString("hello")),
				},
				Requests:       1,
				Concurrency:    2,
				Interval:       3 * time.Second,
				RequestTimeout: 4 * time.Second,
				GlobalTimeout:  5 * time.Second,
			},
			b.Runner(),
		)
	})
}

func mustParseURL(v string) *url.URL {
	u, err := url.ParseRequestURI(v)
	if err != nil {
		panic("mustParseURL: " + err.Error())
	}
	return u
}
