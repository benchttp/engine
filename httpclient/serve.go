package httpclient

import (
	"fmt"
	"net/http"
)

func ListenAndServe(port string) error {
	return (&http.Server{
		Addr:    ":" + fmt.Sprint(port),
		Handler: http.HandlerFunc(handle),
		// No timeout because the end user runs both the client and
		// the server on their machine, which makes it irrelevant.
		ReadHeaderTimeout: 0,
	}).ListenAndServe()
}
