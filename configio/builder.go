package configio

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

// A Builder is used to incrementally build a benchttp.Runner
// using Set and Write methods.
// The zero value is ready to use.
type Builder struct {
	mutations []func(*benchttp.Runner)
}

// WriteJSON decodes the input bytes as a JSON benchttp configuration
// and appends the resulting modifier to the builder.
// It returns any encountered during the decoding process or if the
// decoded configuration is invalid.
func (b *Builder) WriteJSON(in []byte) error {
	return b.decodeAndWrite(in, FormatJSON)
}

// WriteYAML decodes the input bytes as a YAML benchttp configuration
// and appends the resulting modifier to the builder.
// It returns any encountered during the decoding process or if the
// decoded configuration is invalid.
func (b *Builder) WriteYAML(in []byte) error {
	return b.decodeAndWrite(in, FormatYAML)
}

func (b *Builder) decodeAndWrite(in []byte, format Format) error {
	repr := representation{}
	if err := decoderOf(format, in).decodeRepr(&repr); err != nil {
		return err
	}
	// early check for invalid configuration
	if err := repr.validate(); err != nil {
		return err
	}
	b.append(func(dst *benchttp.Runner) {
		// err is already checked via repr.validate(), so nil is expected.
		if err := repr.parseAndMutate(dst); err != nil {
			panicInternal("Builder.decodeAndWrite", "unexpected error: "+err.Error())
		}
	})
	return nil
}

// Runner successively applies the Builder's mutations
// to a zero benchttp.Runner and returns it.
func (b *Builder) Runner() benchttp.Runner {
	runner := benchttp.Runner{}
	b.Mutate(&runner)
	return runner
}

// Mutate successively applies the Builder's mutations
// to the benchttp.Runner value pointed to by dst.
func (b *Builder) Mutate(dst *benchttp.Runner) {
	for _, mutate := range b.mutations {
		mutate(dst)
	}
}

// setters

// SetRequest adds a mutation that sets a runner's request to r.
func (b *Builder) SetRequest(r *http.Request) {
	b.append(func(runner *benchttp.Runner) {
		runner.Request = r
	})
}

// SetRequestMethod adds a mutation that sets a runner's request method to v.
func (b *Builder) SetRequestMethod(v string) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Method = v
	})
}

// SetRequestURL adds a mutation that sets a runner's request URL to v.
func (b *Builder) SetRequestURL(v *url.URL) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.URL = v
	})
}

// SetRequestHeader adds a mutation that sets a runner's request header to v.
func (b *Builder) SetRequestHeader(v http.Header) {
	b.SetRequestHeaderFunc(func(_ http.Header) http.Header {
		return v
	})
}

// SetRequestHeaderFunc adds a mutation that sets a runner's request header
// to the result of calling f with its current request header.
func (b *Builder) SetRequestHeaderFunc(f func(prev http.Header) http.Header) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Header = f(runner.Request.Header)
	})
}

// SetRequestBody adds a mutation that sets a runner's request body to v.
func (b *Builder) SetRequestBody(v io.ReadCloser) {
	b.append(func(runner *benchttp.Runner) {
		if runner.Request == nil {
			runner.Request = &http.Request{}
		}
		runner.Request.Body = v
	})
}

// SetRequests adds a mutation that sets a runner's
// Requests field to v.
func (b *Builder) SetRequests(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Requests = v
	})
}

// SetConcurrency adds a mutation that sets a runner's
// Concurrency field to v.
func (b *Builder) SetConcurrency(v int) {
	b.append(func(runner *benchttp.Runner) {
		runner.Concurrency = v
	})
}

// SetInterval adds a mutation that sets a runner's
// Interval field to v.
func (b *Builder) SetInterval(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.Interval = v
	})
}

// SetRequestTimeout adds a mutation that sets a runner's
// RequestTimeout field to v.
func (b *Builder) SetRequestTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.RequestTimeout = v
	})
}

// SetGlobalTimeout adds a mutation that sets a runner's
// GlobalTimeout field to v.
func (b *Builder) SetGlobalTimeout(v time.Duration) {
	b.append(func(runner *benchttp.Runner) {
		runner.GlobalTimeout = v
	})
}

// SetTests adds a mutation that sets a runner's
// Tests field to v.
func (b *Builder) SetTests(v []benchttp.TestCase) {
	b.append(func(runner *benchttp.Runner) {
		runner.Tests = v
	})
}

// SetTests adds a mutation that appends the given benchttp.TestCases
// to a runner's Tests field.
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
