package configflags

import (
	"flag"
	"net/http"
	"net/url"

	"github.com/benchttp/engine/runner"
)

// Bind reads arguments provided to flagset as config.Fields and binds
// their value to the appropriate fields of given *config.Global.
// The provided *flag.Flagset must not have been parsed yet, otherwise
// bindings its values would fail.
func Bind(flagset *flag.FlagSet, dst *runner.Config) {
	// avoid nil pointer dereferences
	if dst.Request.URL == nil {
		dst.Request.URL = &url.URL{}
	}
	if dst.Request.Header == nil {
		dst.Request.Header = http.Header{}
	}

	// request url
	flagset.Var(urlValue{url: dst.Request.URL},
		runner.ConfigFieldURL,
		runner.ConfigFieldsUsage[runner.ConfigFieldURL],
	)
	// request method
	flagset.StringVar(&dst.Request.Method,
		runner.ConfigFieldMethod,
		dst.Request.Method,
		runner.ConfigFieldsUsage[runner.ConfigFieldMethod],
	)
	// request header
	flagset.Var(headerValue{header: &dst.Request.Header},
		runner.ConfigFieldHeader,
		runner.ConfigFieldsUsage[runner.ConfigFieldHeader],
	)
	// request body
	flagset.Var(bodyValue{body: &dst.Request.Body},
		runner.ConfigFieldBody,
		runner.ConfigFieldsUsage[runner.ConfigFieldBody],
	)
	// requests number
	flagset.IntVar(&dst.Runner.Requests,
		runner.ConfigFieldRequests,
		dst.Runner.Requests,
		runner.ConfigFieldsUsage[runner.ConfigFieldRequests],
	)

	// concurrency
	flagset.IntVar(&dst.Runner.Concurrency,
		runner.ConfigFieldConcurrency,
		dst.Runner.Concurrency,
		runner.ConfigFieldsUsage[runner.ConfigFieldConcurrency],
	)
	// non-conurrent requests interval
	flagset.DurationVar(&dst.Runner.Interval,
		runner.ConfigFieldInterval,
		dst.Runner.Interval,
		runner.ConfigFieldsUsage[runner.ConfigFieldInterval],
	)
	// request timeout
	flagset.DurationVar(&dst.Runner.RequestTimeout,
		runner.ConfigFieldRequestTimeout,
		dst.Runner.RequestTimeout,
		runner.ConfigFieldsUsage[runner.ConfigFieldRequestTimeout],
	)
	// global timeout
	flagset.DurationVar(&dst.Runner.GlobalTimeout,
		runner.ConfigFieldGlobalTimeout,
		dst.Runner.GlobalTimeout,
		runner.ConfigFieldsUsage[runner.ConfigFieldGlobalTimeout],
	)

	// silent mode
	flagset.BoolVar(&dst.Output.Silent,
		runner.ConfigFieldSilent,
		dst.Output.Silent,
		runner.ConfigFieldsUsage[runner.ConfigFieldSilent],
	)
	// output template
	flagset.StringVar(&dst.Output.Template,
		runner.ConfigFieldTemplate,
		dst.Output.Template,
		runner.ConfigFieldsUsage[runner.ConfigFieldTemplate],
	)
}
