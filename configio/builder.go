package configio

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

type Builder struct {
	// TODO: benchmark this vs []func(*benchttp.Runner)
	modifier func(*benchttp.Runner)
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
	b.pipe(func(dst *benchttp.Runner) {
		_ = repr.Into(dst)
	})
	return nil
}

func (b *Builder) Runner() benchttp.Runner {
	runner := benchttp.Runner{}
	b.into(&runner)
	return runner
}

func (b *Builder) Into(dst *benchttp.Runner) {
	b.into(dst)
}

func (b *Builder) into(dst *benchttp.Runner) {
	if b.modifier == nil {
		return
	}
	b.modifier(dst)
}

// setters

func (b *Builder) SetRequest(r *http.Request) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Request = r
	})
}

func (b *Builder) SetRequestMethod(v string) {
	b.pipe(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Method = v
	})
}

func (b *Builder) SetRequestURL(v *url.URL) {
	b.pipe(func(runner *benchttp.Runner) {
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
	b.pipe(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Header = f(runner.Request.Header)
	})
}

func (b *Builder) SetRequestBody(v io.ReadCloser) {
	b.pipe(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Body = v
	})
}

func (b *Builder) SetRequests(v int) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Requests = v
	})
}

func (b *Builder) SetConcurrency(v int) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Concurrency = v
	})
}

func (b *Builder) SetInterval(v time.Duration) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Interval = v
	})
}

func (b *Builder) SetRequestTimeout(v time.Duration) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.RequestTimeout = v
	})
}

func (b *Builder) SetGlobalTimeout(v time.Duration) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.GlobalTimeout = v
	})
}

func (b *Builder) SetTests(v []benchttp.TestCase) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Tests = v
	})
}

func (b *Builder) AddTests(v ...benchttp.TestCase) {
	b.pipe(func(runner *benchttp.Runner) {
		runner.Tests = append(runner.Tests, v...)
	})
}

func (b *Builder) pipe(modifier func(runner *benchttp.Runner)) {
	if modifier == nil {
		panicInternal("Builder.pipe", "call with nil modifier")
	}
	if b.modifier == nil {
		b.modifier = modifier
		return
	}
	applyPreviousModifier := b.modifier
	b.modifier = func(dst *benchttp.Runner) {
		applyPreviousModifier(dst)
		modifier(dst)
	}
}
