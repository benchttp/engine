package requester

import (
	"io"
	"net/http"
	"time"
)

// newClient returns a new http.Client with the given transport and timeout.
func newClient(transport http.RoundTripper, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// cloneRequest fully clones a *http.Request by also cloning the body
// via Request.GetBody.
func cloneRequest(req *http.Request) *http.Request {
	reqClone := req.Clone(req.Context())
	if req.Body != nil {
		// err is always nil (https://golang.org/src/net/http/request.go#L889)
		reqClone.Body, _ = req.GetBody()
	}
	return reqClone
}

// readClose reads resp.Body and closes it.
func readClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
