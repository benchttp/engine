package configflags

import (
	"flag"
	"net/http"
	"net/url"

	"github.com/benchttp/runner/config"
)

// Set reads arguments provided to flagset as config.Fields and binds
// their value to the appropriate fields of given *config.Global.
// The provided *flag.Flagset must not have been parsed yet, otherwise
// bindings its values would fail.
func Set(flagset *flag.FlagSet, dst *config.Global) {
	// avoid nil pointer dereferences
	if dst.Request.URL == nil {
		dst.Request.URL = &url.URL{}
	}
	if dst.Request.Header == nil {
		dst.Request.Header = http.Header{}
	}

	// request url
	flagset.Var(urlValue{url: dst.Request.URL},
		config.FieldURL,
		config.FieldsUsage[config.FieldURL],
	)
	// request method
	flagset.StringVar(&dst.Request.Method,
		config.FieldMethod,
		dst.Request.Method,
		config.FieldsUsage[config.FieldMethod],
	)
	// request header
	flagset.Var(headerValue{header: &dst.Request.Header},
		config.FieldHeader,
		config.FieldsUsage[config.FieldHeader],
	)
	// request body
	flagset.Var(bodyValue{body: &dst.Request.Body},
		config.FieldBody,
		config.FieldsUsage[config.FieldBody],
	)
	// requests number
	flagset.IntVar(&dst.Runner.Requests,
		config.FieldRequests,
		dst.Runner.Requests,
		config.FieldsUsage[config.FieldRequests],
	)

	// concurrency
	flagset.IntVar(&dst.Runner.Concurrency,
		config.FieldConcurrency,
		dst.Runner.Concurrency,
		config.FieldsUsage[config.FieldConcurrency],
	)
	// non-conurrent requests interval
	flagset.DurationVar(&dst.Runner.Interval,
		config.FieldInterval,
		dst.Runner.Interval,
		config.FieldsUsage[config.FieldInterval],
	)
	// request timeout
	flagset.DurationVar(&dst.Runner.RequestTimeout,
		config.FieldRequestTimeout,
		dst.Runner.RequestTimeout,
		config.FieldsUsage[config.FieldRequestTimeout],
	)
	// global timeout
	flagset.DurationVar(&dst.Runner.GlobalTimeout,
		config.FieldGlobalTimeout,
		dst.Runner.GlobalTimeout,
		config.FieldsUsage[config.FieldGlobalTimeout],
	)

	// output strategies
	flagset.Var(outValue{out: &dst.Output.Out},
		config.FieldOut,
		config.FieldsUsage[config.FieldOut],
	)
	// silent mode
	flagset.BoolVar(&dst.Output.Silent,
		config.FieldSilent,
		dst.Output.Silent,
		config.FieldsUsage[config.FieldSilent],
	)
	// output template
	flagset.StringVar(&dst.Output.Template,
		config.FieldTemplate,
		dst.Output.Template,
		config.FieldsUsage[config.FieldTemplate],
	)
}
