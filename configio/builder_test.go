package configio_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/benchttptest"
	"github.com/benchttp/sdk/configio"
)

func TestBuilder_WriteJSON(t *testing.T) {
	in := []byte(`{"runner":{"requests": 5}}`)
	dest := benchttp.Runner{Requests: 0, Concurrency: 2}
	want := benchttp.Runner{Requests: 5, Concurrency: 2}

	b := configio.Builder{}
	if err := b.DecodeJSON(in); err != nil {
		t.Fatal(err)
	}
	b.Mutate(&dest)

	benchttptest.AssertEqualRunners(t, want, dest)
}

func TestBuilder_WriteYAML(t *testing.T) {
	in := []byte(`runner: { requests: 5 }`)
	dest := benchttp.Runner{Requests: 0, Concurrency: 2}
	want := benchttp.Runner{Requests: 5, Concurrency: 2}

	b := configio.Builder{}
	if err := b.DecodeYAML(in); err != nil {
		t.Fatal(err)
	}
	b.Mutate(&dest)

	benchttptest.AssertEqualRunners(t, want, dest)
}

func TestBuilder_Set(t *testing.T) {
	t.Run("basic fields", func(t *testing.T) {
		want := benchttp.Runner{
			Requests:       5,
			Concurrency:    2,
			Interval:       10 * time.Millisecond,
			RequestTimeout: 1 * time.Second,
			GlobalTimeout:  10 * time.Second,
		}

		b := configio.Builder{}
		b.SetRequests(want.Requests)
		b.SetConcurrency(-1)
		b.SetConcurrency(want.Concurrency)
		b.SetInterval(want.Interval)
		b.SetRequestTimeout(want.RequestTimeout)
		b.SetGlobalTimeout(want.GlobalTimeout)

		benchttptest.AssertEqualRunners(t, want, b.Runner())
	})

	t.Run("request", func(t *testing.T) {
		want := benchttp.Runner{
			Request: httptest.NewRequest("GET", "https://example.com", nil),
		}

		b := configio.Builder{}
		b.SetRequest(want.Request)

		benchttptest.AssertEqualRunners(t, want, b.Runner())
	})

	t.Run("request fields", func(t *testing.T) {
		want := benchttp.Runner{
			Request: &http.Request{
				Method: "PUT",
				URL:    mustParseRequestURI("https://example.com"),
				Header: http.Header{
					"API_KEY": []string{"abc"},
					"Accept":  []string{"text/html", "application/json"},
				},
				Body: readcloser("hello"),
			},
		}

		b := configio.Builder{}
		b.SetRequestMethod(want.Request.Method)
		b.SetRequestURL(want.Request.URL)
		b.SetRequestHeader(http.Header{"API_KEY": []string{"abc"}})
		b.SetRequestHeaderFunc(func(prev http.Header) http.Header {
			prev.Add("Accept", "text/html")
			prev.Add("Accept", "application/json")
			return prev
		})
		b.SetRequestBody(readcloser("hello"))

		benchttptest.AssertEqualRunners(t, want, b.Runner())
	})

	t.Run("test cases", func(t *testing.T) {
		want := benchttp.Runner{
			Tests: []benchttp.TestCase{
				{
					Name:      "maximum response time",
					Field:     "ResponseTimes.Max",
					Predicate: "LT",
					Target:    100 * time.Millisecond,
				},
				{
					Name:      "similar response times",
					Field:     "ResponseTimes.StdDev",
					Predicate: "LTE",
					Target:    20 * time.Millisecond,
				},
				{
					Name:      "100% availability",
					Field:     "RequestFailureCount",
					Predicate: "EQ",
					Target:    0,
				},
			},
		}

		b := configio.Builder{}
		b.SetTests([]benchttp.TestCase{want.Tests[0]})
		b.AddTests(want.Tests[1:]...)

		benchttptest.AssertEqualRunners(t, want, b.Runner())
	})
}

// helpers

func mustParseRequestURI(s string) *url.URL {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		panic("mustParseRequestURI: " + err.Error())
	}
	return u
}

func readcloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
