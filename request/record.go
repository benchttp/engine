package request

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Record struct {
	cost  time.Duration
	code  int
	bytes int
	error string
}

func newRecord(resp *http.Response, t time.Duration) Record {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	r := Record{
		code:  resp.StatusCode,
		cost:  t,
		bytes: len(body),
	}

	if err != nil {
		r.error = fmt.Sprint(err)
	}

	return r
}

func (r Record) String() string {
	return fmt.Sprintf("took %s", r.cost)
}
