package httpclient

import (
	"fmt"
	"net/http"
)

// ListenAndServe listen to the given port and serves routes
// to run and stream a benchttp benchmark.
func ListenAndServe(port string) error {
	return (&http.Server{
		Addr:    ":" + fmt.Sprint(port),
		Handler: http.HandlerFunc(handle),
		// No timeout because the end user runs both the client and
		// the server on their machine, which makes it irrelevant.
		ReadHeaderTimeout: 0,
	}).ListenAndServe()
}
