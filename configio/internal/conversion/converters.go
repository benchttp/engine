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
	encode func(src benchttp.Runner, dst *Repr)
	decode func(src Repr, dst *benchttp.Runner) error
}

type requestConverter struct {
	encode func(src *http.Request, dst *Repr)
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
	encode: func(src benchttp.Runner, dst *Repr) {
		if src.Request != nil {
			for _, c := range requestConverters {
				c.encode(src.Request, dst)
			}
		}
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
	encode: func(src benchttp.Runner, dst *Repr) {
		for _, c := range runnerConverters {
			c.encode(src, dst)
		}
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
	encode: func(src benchttp.Runner, dst *Repr) {
		for _, c := range src.Tests {
			// /!\ loop ref hazard
			name := c.Name
			field := string(c.Field)
			predicate := string(c.Predicate)
			target := c.Target
			switch t := target.(type) {
			case time.Duration:
				target = t.String()
			}
			dst.Tests = append(dst.Tests, testcaseRepr{
				Name:      &name,
				Field:     &field,
				Predicate: &predicate,
				Target:    &target,
			})
		}
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

var fieldRequestMethod = requestConverter{
	decode: func(src Repr, dst *http.Request) error {
		if m := src.Request.Method; m != nil {
			dst.Method = *m
		}
		return nil
	},
	encode: func(src *http.Request, dst *Repr) {
		dst.Request.Method = &src.Method
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
	encode: func(src *http.Request, dst *Repr) {
		s := src.URL.String()
		dst.Request.URL = &s
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
	encode: func(src *http.Request, dst *Repr) {
		dst.Request.Header = src.Header
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
	encode: func(src *http.Request, dst *Repr) {
		// TODO
	},
}

var fieldRunnerRequests = bindInt(
	func(src *Repr, dst *benchttp.Runner) (*int, *int) {
		return src.Runner.Requests, &dst.Requests
	},
)

var fieldRunnerConcurrency = bindInt(
	func(src *Repr, dst *benchttp.Runner) (*int, *int) {
		return src.Runner.Concurrency, &dst.Concurrency
	},
)

var fieldRunnerInterval = bindDuration(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.Interval, &dst.Interval
	},
)

var fieldRunnerRequestTimeout = bindDuration(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.RequestTimeout, &dst.RequestTimeout
	},
)

var fieldRunnerGlobalTimeout = bindDuration(
	func(src *Repr, dst *benchttp.Runner) (*string, *time.Duration) {
		return src.Runner.GlobalTimeout, &dst.GlobalTimeout
	},
)

func bindDuration(
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
		encode: func(src benchttp.Runner, dst *Repr) {
			if vdst, vsrc := bind(dst, &src); vsrc != nil {
				// FIXME: nil pointer deref
				*vdst = vsrc.String()
			}
		},
	}
}

func bindInt(
	bind func(*Repr, *benchttp.Runner) (*int, *int),
) converter {
	return converter{
		decode: func(src Repr, dst *benchttp.Runner) error {
			if vsrc, vdst := bind(&src, dst); vsrc != nil {
				*vdst = *vsrc
			}
			return nil
		},
		encode: func(src benchttp.Runner, dst *Repr) {
			if vdst, vsrc := bind(dst, &src); vsrc != nil {
				// FIXME: nil pointer deref
				*vdst = *vsrc
			}
		},
	}
}
