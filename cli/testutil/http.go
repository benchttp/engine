package testutil

import (
	"bytes"
	"net/http"
)

func MustMakeRequest(method, uri string, header http.Header, body []byte) *http.Request {
	req, err := http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		panic("testutil.MustMakeRequest: " + err.Error())
	}
	req.Header = header
	return req
}
