package benchttptest

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/benchttp/engine/benchttp"
)

// RunnerCmpOptions is the cmp.Options used to compare benchttp.Runner.
// By default, it ignores unexported fields and includes RequestCmpOptions.
var RunnerCmpOptions = cmp.Options{
	cmpopts.IgnoreUnexported(benchttp.Runner{}),
	RequestCmpOptions,
}

// RequestCmpOptions is the cmp.Options used to compare *http.Request.
// It behaves as follows:
//
//   - Nil and empty values are considered equal
//
//   - Fields that depend on how the request was created are ignored
//     to avoid false negatives when comparing requests created in different
//     ways (http.NewRequest vs httptest.NewRequest vs &http.Request{})
//
//   - Function fields are ignored
//
//   - Body is ignored: it is compared separately
var RequestCmpOptions = cmp.Options{
	cmp.Transformer("Request", instantiateNilRequest),
	cmp.Transformer("Request.Header", instantiateNilHeader),
	cmp.Transformer("Request.URL", stringifyURL),
	cmpopts.IgnoreUnexported(http.Request{}, tls.ConnectionState{}),
	cmpopts.IgnoreFields(http.Request{}, unreliableRequestFields...),
}

var unreliableRequestFields = []string{
	// These fields are automatically set by NewRequest constructor
	// from packages http and httptest, as a consequence they can
	// trigger false positives when comparing requests that were
	// created differently.
	"Proto", "ProtoMajor", "ProtoMinor", "ContentLength",
	"Host", "RemoteAddr", "RequestURI", "TLS", "Cancel",

	// Function fields cannot be reliably compared
	"GetBody",

	// Body field can't be read without altering the Request, causing
	// cmp-go to panic. We perform a custom comparison instead.
	"Body",
}

// AssertEqualRunners fails t and shows a diff if a and b are not equal,
// as determined by RunnerCmpOptions.
func AssertEqualRunners(t *testing.T, x, y benchttp.Runner) {
	t.Helper()
	if !EqualRunners(x, y) {
		t.Error(DiffRunner(x, y))
	}
}

// EqualRunners returns true if x and y are equal, as determined by
// RunnerCmpOptions.
func EqualRunners(x, y benchttp.Runner) bool {
	return cmp.Equal(x, y, RunnerCmpOptions) &&
		compareRequestBody(x.Request, y.Request)
}

// DiffRunner returns a string showing the diff between x and y,
// as determined by RunnerCmpOptions.
func DiffRunner(x, y benchttp.Runner) string {
	b := strings.Builder{}
	b.WriteString(cmp.Diff(x, y, RunnerCmpOptions))
	if x.Request != nil && y.Request != nil {
		xbody := nopreadBody(x.Request)
		ybody := nopreadBody(y.Request)
		if !bytes.Equal(xbody, ybody) {
			b.WriteString("Request.Body: ")
			b.WriteString(cmp.Diff(string(xbody), string(ybody)))
		}
	}
	return b.String()
}

// helpers

func instantiateNilRequest(r *http.Request) *http.Request {
	if r == nil {
		return &http.Request{}
	}
	return r
}

func instantiateNilHeader(h http.Header) http.Header {
	if h == nil {
		return http.Header{}
	}
	return h
}

func stringifyURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func compareRequestBody(a, b *http.Request) bool {
	ba, bb := nopreadBody(a), nopreadBody(b)
	return bytes.Equal(ba, bb)
}

func nopreadBody(r *http.Request) []byte {
	if r == nil || r.Body == nil {
		return []byte{}
	}

	bbuf := bytes.Buffer{}

	if _, err := io.Copy(&bbuf, r.Body); err != nil {
		panic("benchttptest: error reading Request.Body: " + err.Error())
	}

	if r.GetBody != nil {
		newbody, err := r.GetBody()
		if err != nil {
			panic("benchttptest: Request.GetBody error: " + err.Error())
		}
		r.Body = newbody
	} else {
		r.Body = io.NopCloser(&bbuf)
	}

	return bbuf.Bytes()
}
