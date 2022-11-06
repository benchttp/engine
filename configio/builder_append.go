package configio

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

type Builder_append struct { // nolint:revive
	modifiers []func(*benchttp.Runner)
}

func (b *Builder_append) WriteJSON(in []byte) error {
	return b.decodeAndWrite(in, FormatJSON)
}

func (b *Builder_append) WriteYAML(in []byte) error {
	return b.decodeAndWrite(in, FormatYAML)
}

func (b *Builder_append) decodeAndWrite(in []byte, format Format) error {
	repr := Representation{}
	if err := DecoderOf(format, in).Decode(&repr); err != nil {
		return err
	}
	if err := repr.validate(); err != nil {
		return err
	}
	b.append(func(dst *benchttp.Runner) {
		_ = repr.Into(dst)
	})
	return nil
}

func (b *Builder_append) Runner() benchttp.Runner {
	runner := benchttp.Runner{}
	b.into(&runner)
	return runner
}

func (b *Builder_append) Into(dst *benchttp.Runner) {
	b.into(dst)
}

func (b *Builder_append) into(dst *benchttp.Runner) {
	for _, modify := range b.modifiers {
		modify(dst)
	}
}

// setters

func (b *Builder_append) SetRequest(r *http.Request) {
	b.append(func(runner *benchttp.Runner) {
		runner.Request = r
	})
}

func (b *Builder_append) SetRequestMethod(v string) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Method = v
	})
}

func (b *Builder_append) SetRequestURL(v *url.URL) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.URL = v
	})
}

func (b *Builder_append) SetRequestHeader(v http.Header) {
	b.SetRequestHeaderFunc(func(_ http.Header) http.Header {
		return v
	})
}

func (b *Builder_append) SetRequestHeaderFunc(f func(prev http.Header) http.Header) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Header = f(runner.Request.Header)
	})
}

func (b *Builder_append) SetRequestBody(v io.ReadCloser) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Body = v
	})
}

func (b *Builder_append) SetRequests(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Requests = v
	})
}

func (b *Builder_append) SetConcurrency(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Concurrency = v
	})
}

func (b *Builder_append) SetInterval(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.Interval = v
	})
}

func (b *Builder_append) SetRequestTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.RequestTimeout = v
	})
}

func (b *Builder_append) SetGlobalTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.GlobalTimeout = v
	})
}

func (b *Builder_append) SetTests(v []benchttp.TestCase) {
	b.append(func(runner *benchttp.Runner) {
		runner.Tests = v
	})
}

func (b *Builder_append) AddTests(v ...benchttp.TestCase) {
	b.append(func(runner *benchttp.Runner) {
		runner.Tests = append(runner.Tests, v...)
	})
}

func (b *Builder_append) append(modifier func(runner *benchttp.Runner)) {
	if modifier == nil {
		panicInternal("Builder.append", "call with nil modifier")
	}
	b.modifiers = append(b.modifiers, modifier)
}
