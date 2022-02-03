package config

type ErrInvalid struct {
	invalidValues []error
}

func (e *ErrInvalid) Error() string {
	message := "Invalid value(s) provided:\n"
	for _, err := range e.invalidValues {
		message += err.Error() + "\n"
	}
	return message
}
