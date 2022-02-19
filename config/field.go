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

func IsField(v string) bool {
	switch v {
	case FieldMethod, FieldURL, FieldHeader, FieldBody, FieldRequests,
		FieldConcurrency, FieldInterval, FieldRequestTimeout,
		FieldGlobalTimeout, FieldOut, FieldSilent, FieldTemplate:
		return true
	}
	return false
}
