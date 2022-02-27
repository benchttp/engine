package export_test

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/benchttp/runner/output/export"
)

func TestHTTP(t *testing.T) {
	httpDefaultClient := *http.DefaultClient
	resetHTTPDefaultClient := func() {
		*http.DefaultClient = httpDefaultClient
	}

	for _, tc := range []struct {
		label         string
		expErrName    string
		expErr        error
		httpRequester export.HTTPRequester
		httpTransport http.RoundTripper
	}{
		{
			label:         "return ErrHTTPRequest on src.HTTPRequest error",
			expErrName:    "ErrHTTPRequest",
			expErr:        export.ErrHTTPRequest,
			httpRequester: mockRequester{valid: false},
			httpTransport: mockTransport{valid: true, code: 200},
		},
		{
			label:         "return ErrHTTPConnection on HTTP connection error",
			expErrName:    "ErrHTTPConnection",
			expErr:        export.ErrHTTPConnection,
			httpRequester: mockRequester{valid: true},
			httpTransport: mockTransport{valid: false},
		},
		{
			label:         "return ErrHTTPResponse on bad status code",
			expErrName:    "ErrHTTPResponse",
			expErr:        export.ErrHTTPResponse,
			httpRequester: mockRequester{valid: true},
			httpTransport: mockTransport{valid: true, code: 400},
		},
		{
			label:         "happy path",
			expErrName:    "nil",
			expErr:        nil,
			httpRequester: mockRequester{valid: true},
			httpTransport: mockTransport{valid: true, code: 200},
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			t.Cleanup(resetHTTPDefaultClient)
			*http.DefaultClient = *newClientWithTransport(tc.httpTransport)

			gotErr := export.HTTP(tc.httpRequester)
			if !errors.Is(gotErr, tc.expErr) {
				t.Errorf("unexpected error:\nexp %v\ngot %v", tc.expErrName, gotErr)
			}
		})
	}
}

type mockRequester struct{ valid bool }

func (r mockRequester) HTTPRequest() (*http.Request, error) {
	if !r.valid {
		return nil, errors.New("bad request")
	}
	return http.NewRequest("POST", "https://a.b", bytes.NewReader([]byte("abc")))
}

type mockTransport struct {
	code  int
	valid bool
}

func (tr mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !tr.valid {
		return nil, errors.New("connection error")
	}
	return &http.Response{StatusCode: tr.code}, nil
}

func newClientWithTransport(tr http.RoundTripper) *http.Client {
	return &http.Client{Transport: tr}
}
