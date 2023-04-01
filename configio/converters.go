package configio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/engine/benchttp"
)

type (
	converter        func(representation, *benchttp.Runner) error
	requestConverter func(representation, *http.Request) error
)

var fieldConverters = []converter{
	func(repr representation, dst *benchttp.Runner) error {
		req := dst.Request
		if req == nil {
			req = &http.Request{}
		}
		for _, p := range requestConverters {
			if err := p(repr, req); err != nil {
				return err
			}
		}
		dst.Request = req
		return nil
	},
	func(repr representation, dst *benchttp.Runner) error {
		for _, p := range runnerParsers {
			if err := p(repr, dst); err != nil {
				return err
			}
		}
		return nil
	},
}

var requestParser = func(repr representation, dst *benchttp.Runner) error {
	req := &http.Request{}
	for _, fieldParser := range requestConverters {
		if err := fieldParser(repr, req); err != nil {
			return err
		}
	}
	dst.Request = req
	return nil
}

var requestConverters = []requestConverter{
	func(repr representation, dst *http.Request) error {
		return setString(repr.Request.Method, &dst.Method)
	},
	func(repr representation, dst *http.Request) error {
		if rawURL := repr.Request.URL; rawURL != nil {
			parsedURL, err := parseAndBuildURL(*rawURL, repr.Request.QueryParams)
			if err != nil {
				return fmt.Errorf(`configio: invalid url: %q`, *rawURL)
			}
			dst.URL = parsedURL
		}
		return nil
	},
	func(repr representation, dst *http.Request) error {
		if header := repr.Request.Header; len(header) != 0 {
			httpHeader := http.Header{}
			for key, val := range header {
				httpHeader[key] = val
			}
			dst.Header = httpHeader
		}
		return nil
	},
	func(repr representation, dst *http.Request) error {
		if body := repr.Request.Body; body != nil {
			switch body.Type {
			case "raw":
				dst.Body = io.NopCloser(bytes.NewReader([]byte(body.Content)))
			default:
				return errors.New(`configio: request.body.type: only "raw" accepted`)
			}
		}
		return nil
	},
}

var runnerParsers = map[string]converter{
	"requests": func(repr representation, dst *benchttp.Runner) error {
		return setInt(repr.Runner.Requests, &dst.Requests)
	},
	"concurrency": func(repr representation, dst *benchttp.Runner) error {
		return setInt(repr.Runner.Concurrency, &dst.Concurrency)
	},
	"interval": func(repr representation, dst *benchttp.Runner) error {
		return setOptionalDuration(repr.Runner.Interval, &dst.Interval)
	},
	"requestTimeout": func(repr representation, dst *benchttp.Runner) error {
		return setOptionalDuration(repr.Runner.RequestTimeout, &dst.RequestTimeout)
	},
	"globalTimeout": func(repr representation, dst *benchttp.Runner) error {
		return setOptionalDuration(repr.Runner.GlobalTimeout, &dst.GlobalTimeout)
	},
}

func setInt(src, dst *int) error {
	if src != nil {
		*dst = *src
	}
	return nil
}

func setString(src, dst *string) error {
	if src != nil {
		*dst = *src
	}
	return nil
}

func setOptionalDuration(src *string, dst *time.Duration) error {
	if src == nil {
		return nil
	}
	parsed, err := parseOptionalDuration(*src)
	if err != nil {
		return err
	}
	*dst = parsed
	return nil
}
