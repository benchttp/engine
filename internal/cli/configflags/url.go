package configflags

import (
	"fmt"
	"net/url"
)

// urlValue implements flag.Value
type urlValue struct {
	url *url.URL
}

// String returns a string representation of urlValue.url.
func (v urlValue) String() string {
	if v.url == nil {
		return ""
	}
	return v.url.String()
}

// Set parses input string as a URL and sets the referenced URL accordingly.
func (v urlValue) Set(in string) error {
	urlURL, err := url.ParseRequestURI(in)
	if err != nil {
		return fmt.Errorf(`invalid url: "%s"`, in)
	}
	*v.url = *urlURL
	return nil
}
