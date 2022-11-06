package configio

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

type Builder struct {
	mutations []func(*benchttp.Runner)
}

func (b *Builder) WriteJSON(in []byte) error {
	return b.decodeAndWrite(in, FormatJSON)
}

func (b *Builder) WriteYAML(in []byte) error {
	return b.decodeAndWrite(in, FormatYAML)
}

func (b *Builder) decodeAndWrite(in []byte, format Format) error {
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

func (b *Builder) Runner() benchttp.Runner {
	runner := benchttp.Runner{}
	b.Mutate(&runner)
	return runner
}

func (b *Builder) Mutate(dst *benchttp.Runner) {
	for _, mutate := range b.mutations {
		mutate(dst)
	}
}

// setters

func (b *Builder) SetRequest(r *http.Request) {
	b.append(func(runner *benchttp.Runner) {
		runner.Request = r
	})
}

func (b *Builder) SetRequestMethod(v string) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Method = v
	})
}

func (b *Builder) SetRequestURL(v *url.URL) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.URL = v
	})
}

func (b *Builder) SetRequestHeader(v http.Header) {
	b.SetRequestHeaderFunc(func(_ http.Header) http.Header {
		return v
	})
}

func (b *Builder) SetRequestHeaderFunc(f func(prev http.Header) http.Header) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Header = f(runner.Request.Header)
	})
}

func (b *Builder) SetRequestBody(v io.ReadCloser) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Body = v
	})
}

func (b *Builder) SetRequests(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Requests = v
	})
}

func (b *Builder) SetConcurrency(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Concurrency = v
	})
}

func (b *Builder) SetInterval(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.Interval = v
	})
}

func (b *Builder) SetRequestTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.RequestTimeout = v
	})
}

func (b *Builder) SetGlobalTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.GlobalTimeout = v
	})
}

func (b *Builder) SetTests(v []benchttp.TestCase) {
	b.append(func(runner *benchttp.Runner) {
		runner.Tests = v
	})
}

func (b *Builder) AddTests(v ...benchttp.TestCase) {
	b.append(func(runner *benchttp.Runner) {
		runner.Tests = append(runner.Tests, v...)
	})
}

func (b *Builder) append(modifier func(runner *benchttp.Runner)) {
	if modifier == nil {
		panicInternal("Builder.append", "call with nil modifier")
	}
	b.mutations = append(b.mutations, modifier)
}