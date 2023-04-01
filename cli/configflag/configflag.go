package configflag

const (
	flagMethod         = "method"
	flagURL            = "url"
	flagHeader         = "header"
	flagBody           = "body"
	flagRequests       = "requests"
	flagConcurrency    = "concurrency"
	flagInterval       = "interval"
	flagRequestTimeout = "requestTimeout"
	flagGlobalTimeout  = "globalTimeout"
	flagTests          = "tests"
)

// flagsUsage is a record of all available config flags and their usage.
var flagsUsage = map[string]string{
	flagMethod:         "HTTP request method",
	flagURL:            "HTTP request url",
	flagHeader:         "HTTP request header",
	flagBody:           "HTTP request body",
	flagRequests:       "Number of requests to run, use duration as exit condition if omitted",
	flagConcurrency:    "Number of connections to run concurrently",
	flagInterval:       "Minimum duration between two non concurrent requests",
	flagRequestTimeout: "Timeout for each HTTP request",
	flagGlobalTimeout:  "Max duration of test",
	flagTests:          "Test suite",
}
