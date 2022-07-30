package runner

const (
	ConfigFieldMethod         = "method"
	ConfigFieldURL            = "url"
	ConfigFieldHeader         = "header"
	ConfigFieldBody           = "body"
	ConfigFieldRequests       = "requests"
	ConfigFieldConcurrency    = "concurrency"
	ConfigFieldInterval       = "interval"
	ConfigFieldRequestTimeout = "requestTimeout"
	ConfigFieldGlobalTimeout  = "globalTimeout"
	ConfigFieldSilent         = "silent"
	ConfigFieldTemplate       = "template"
)

// ConfigFieldsUsage is a record of all available config fields and their usage.
var ConfigFieldsUsage = map[string]string{
	ConfigFieldMethod:         "HTTP request method",
	ConfigFieldURL:            "HTTP request url",
	ConfigFieldHeader:         "HTTP request header",
	ConfigFieldBody:           "HTTP request body",
	ConfigFieldRequests:       "Number of requests to run, use duration as exit condition if omitted",
	ConfigFieldConcurrency:    "Number of connections to run concurrently",
	ConfigFieldInterval:       "Minimum duration between two non concurrent requests",
	ConfigFieldRequestTimeout: "Timeout for each HTTP request",
	ConfigFieldGlobalTimeout:  "Max duration of test",
	ConfigFieldSilent:         "Silent mode (no write to stdout)",
	ConfigFieldTemplate:       "Output template",
}

func IsConfigField(v string) bool {
	_, exists := ConfigFieldsUsage[v]
	return exists
}
