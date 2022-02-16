package config

const (
	FieldMethod        = "method"
	FieldURL           = "url"
	FieldHeader        = "header"
	FieldTimeout       = "timeout"
	FieldRequests      = "requests"
	FieldConcurrency   = "concurrency"
	FieldInterval      = "interval"
	FieldGlobalTimeout = "globalTimeout"
	FieldBody          = "body"
)

func IsField(v string) bool {
	switch v {
	case FieldMethod, FieldURL, FieldHeader, FieldTimeout, FieldRequests,
		FieldConcurrency, FieldInterval, FieldGlobalTimeout, FieldBody:
		return true
	}
	return false
}
