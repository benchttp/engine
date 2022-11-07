package conversion

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

type converter struct {
	// TODO:
	// encode func(src benchttp.Runner, dst *Repr)
	decode func(src Repr, dst *benchttp.Runner) error
}

type requestConverter struct {
	// TODO:
	// encode func(src *http.Request, dst *Repr)
	decode func(src Repr, dst *http.Request) error
}

var converters = []converter{
	fieldRequest,
	fieldRunner,
	fieldTests,
}

var fieldRequest = converter{
	decode: func(src Repr, dst *benchttp.Runner) error {
		if dst.Request == nil {
			dst.Request = &http.Request{}
		}
		for _, c := range requestConverters {
			if err := c.decode(src, dst.Request); err != nil {
				return err
			}
		}
		return nil
	},
}

var fieldRunner = converter{
	decode: func(src Repr, dst *benchttp.Runner) error {
		for _, c := range runnerConverters {
			if err := c.decode(src, dst); err != nil {
				return err
			}
		}
		return nil
	},
}

var fieldTests = converter{
	decode: func(src Repr, dst *benchttp.Runner) error {
		if tests := src.Tests; tests != nil {
			cases, err := parseTests(tests)
			if err != nil {
				return err
			}
			dst.Tests = cases
			return nil
		}
		return nil
	},
}

var requestConverters = []requestConverter{
	fieldRequestMethod,
	fieldRequestURL,
	fieldRequestHeader,
	fieldRequestBody,
}

var runnerConverters = []converter{
	fieldRunnerRequests,
	fieldRunnerConcurrency,
	fieldRunnerInterval,
	fieldRunnerRequestTimeout,
	fieldRunnerGlobalTimeout,
}

// TODO:
// var testsConverters = []converter{
// 	fieldTests,
// }

var fieldRequestMethod = requestConverter{
	decode: func(src Repr, dst *http.Request) error {
		if m := src.Request.Method; m != nil {
			dst.Method = *m
		}
		return nil
	},
}

var fieldRequestURL = requestConverter{
	decode: func(src Repr, dst *http.Request) error {
		if rawURL := src.Request.URL; rawURL != nil {
			parsedURL, err := parseAndBuildURL(*rawURL, src.Request.QueryParams)
			if err != nil {
				return fmt.Errorf(`configio: invalid url: %q`, *rawURL)
			}
			dst.URL = parsedURL
		}
		return nil
	},
}

var fieldRequestHeader = requestConverter{
	decode: func(src Repr, dst *http.Request) error {
		if header := src.Request.Header; len(header) != 0 {
			httpHeader := http.Header{}
			for key, val := range header {
				httpHeader[key] = val
			}
			dst.Header = httpHeader
		}
		return nil
	},
}

var fieldRequestBody = requestConverter{
	decode: func(src Repr, dst *http.Request) error {
		if body := src.Request.Body; body != nil {
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

var fieldRunnerRequests = intField(
	func(src *Repr, dst *benchttp.Runner) (*int, *int) {
		return src.Runner.Requests, &dst.Requests
	},
)

var fieldRunnerConcurrency = intField(
	func(src *Repr, dst *benchttp.Runner) (*int, *int) {
		return src.Runner.Concurrency, &dst.Concurrency
	},
)

var fieldRunnerInterval = durationField(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.Interval, &dst.Interval
	},
)

var fieldRunnerRequestTimeout = durationField(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.RequestTimeout, &dst.RequestTimeout
	},
)

var fieldRunnerGlobalTimeout = durationField(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.GlobalTimeout, &dst.GlobalTimeout
	},
)

func durationField(
	bind func(*Repr, *benchttp.Runner) (*string, *time.Duration),
) converter {
	return converter{
		decode: func(src Repr, dst *benchttp.Runner) error {
			if vsrc, vdst := bind(&src, dst); vsrc != nil {
				parsed, err := parseOptionalDuration(*vsrc)
				if err != nil {
					return err
				}
				*vdst = parsed
			}
			return nil
		},
	}
}

func intField(
	bind func(*Repr, *benchttp.Runner) (*int, *int),
) converter {
	return converter{
		decode: func(src Repr, dst *benchttp.Runner) error {
			if vsrc, vdst := bind(&src, dst); vsrc != nil {
				*vdst = *vsrc
			}
			return nil
		},
	}
}
