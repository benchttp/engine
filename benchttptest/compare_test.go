package benchttptest_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/benchttptest"
)

func TestAssertEqualRunners(t *testing.T) {
	for _, tc := range []struct {
		name string
		pass bool
		a, b benchttp.Runner
	}{
		{
			name: "pass if runners are equal",
			pass: true,
			a:    benchttp.Runner{Requests: 1},
			b:    benchttp.Runner{Requests: 1},
		},
		{
			name: "fail if runners are not equal",
			pass: false,
			a:    benchttp.Runner{Requests: 1},
			b:    benchttp.Runner{Requests: 2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tt := &testing.T{}

			benchttptest.AssertEqualRunners(tt, tc.a, tc.b)
			if tt.Failed() == tc.pass {
				t.Fail()
			}
		})
	}
}

func TestEqualRunners(t *testing.T) {
	for _, tc := range []struct {
		name string
		want bool
		a, b benchttp.Runner
	}{
		{
			name: "equal runners",
			want: true,
			a:    benchttp.Runner{Requests: 1},
			b:    benchttp.Runner{Requests: 1},
		},
		{
			name: "different runners",
			want: false,
			a:    benchttp.Runner{Requests: 1},
			b:    benchttp.Runner{Requests: 2},
		},
		{
			name: "consider zero requests equal",
			want: true,
			a:    benchttp.Runner{Request: nil},
			b:    benchttp.Runner{Request: &http.Request{}},
		},
		{
			name: "consider zero request headers equal",
			want: true,
			a:    benchttp.Runner{Request: &http.Request{Header: nil}},
			b:    benchttp.Runner{Request: &http.Request{Header: http.Header{}}},
		},
		{
			name: "consider zero request bodies equal",
			want: true,
			a:    benchttp.Runner{Request: &http.Request{Body: nil}},
			b:    benchttp.Runner{Request: &http.Request{Body: http.NoBody}},
		},
		{
			name: "zero request vs non zero request",
			want: false,
			a:    benchttp.Runner{Request: &http.Request{Method: "GET"}},
			b:    benchttp.Runner{Request: nil},
		},
		{
			name: "different request field values",
			want: false,
			a:    benchttp.Runner{Request: &http.Request{Method: "GET"}},
			b:    benchttp.Runner{Request: &http.Request{Method: "POST"}},
		},
		{
			name: "ignore unreliable request fields",
			want: true,
			a: benchttp.Runner{
				Request: httptest.NewRequest( // sets Proto, ContentLength, ...
					"POST",
					"https://example.com",
					nil,
				),
			},
			b: benchttp.Runner{
				Request: &http.Request{
					Method: "POST",
					URL:    mustParseRequestURI("https://example.com"),
				},
			},
		},
		{
			name: "equal request bodies",
			want: true,
			a: benchttp.Runner{
				Request: &http.Request{
					Body: io.NopCloser(strings.NewReader("hello")),
				},
			},
			b: benchttp.Runner{
				Request: &http.Request{
					Body: io.NopCloser(strings.NewReader("hello")),
				},
			},
		},
		{
			name: "different request bodies",
			want: false,
			a: benchttp.Runner{
				Request: &http.Request{
					Body: io.NopCloser(strings.NewReader("hello")),
				},
			},
			b: benchttp.Runner{
				Request: &http.Request{
					Body: io.NopCloser(strings.NewReader("world")),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if benchttptest.EqualRunners(tc.a, tc.b) != tc.want {
				t.Error(benchttptest.DiffRunner(tc.a, tc.b))
			}
		})
	}

	t.Run("restore request body", func(t *testing.T) {
		a := benchttp.Runner{
			Request: httptest.NewRequest(
				"POST",
				"https://example.com",
				strings.NewReader("hello"),
			),
		}
		b := benchttp.Runner{
			Request: &http.Request{
				Method: "POST",
				URL:    mustParseRequestURI("https://example.com"),
				Body:   io.NopCloser(bytes.NewReader([]byte("hello"))),
			},
		}

		_ = benchttptest.EqualRunners(a, b)

		ba, bb := mustRead(a.Request.Body), mustRead(b.Request.Body)
		want := []byte("hello")
		if !bytes.Equal(want, ba) || !bytes.Equal(want, bb) {
			t.Fail()
		}
	})
}

// helpers

func mustParseRequestURI(s string) *url.URL {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		panic(err)
	}
	return u
}

func mustRead(r io.Reader) []byte {
	b, err := io.ReadAll(r)
	if err != nil {
		panic("mustRead: " + err.Error())
	}
	return b
}
