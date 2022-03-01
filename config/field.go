package config

const (
	FieldMethod         = "method"
	FieldURL            = "url"
	FieldHeader         = "header"
	FieldBody           = "body"
	FieldRequests       = "requests"
	FieldConcurrency    = "concurrency"
	FieldInterval       = "interval"
	FieldRequestTimeout = "requestTimeout"
	FieldGlobalTimeout  = "globalTimeout"
	FieldOut            = "out"
	FieldSilent         = "silent"
	FieldTemplate       = "template"
)

// FieldsUsage is a record of all available config fields and their usage.
var FieldsUsage = map[string]string{
	FieldMethod:         "HTTP request method",
	FieldURL:            "HTTP request url",
	FieldHeader:         "HTTP request header",
	FieldBody:           "HTTP request body",
	FieldRequests:       "Number of requests to run, use duration as exit condition if omitted",
	FieldConcurrency:    "Number of connections to run concurrently",
	FieldInterval:       "Minimum duration between two non concurrent requests",
	FieldRequestTimeout: "Timeout for each HTTP request",
	FieldGlobalTimeout:  "Max duration of test",
	FieldOut:            "Output destination (benchttp,json,stdout)",
	FieldSilent:         "Silent mode (no write to stdout)",
	FieldTemplate:       "Output template",
}

func IsField(v string) bool {
	_, exists := FieldsUsage[v]
	return exists
}
