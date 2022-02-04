package config

const (
	FieldMethod        = "method"
	FieldURL           = "url"
	FieldTimeout       = "timeout"
	FieldRequests      = "requests"
	FieldConcurrency   = "concurrency"
	FieldGlobalTimeout = "globalTimeout"
)

func IsField(v string) bool {
	switch v {
	case FieldMethod, FieldURL, FieldTimeout, FieldRequests,
		FieldConcurrency, FieldGlobalTimeout:
		return true
	}
	return false
}
