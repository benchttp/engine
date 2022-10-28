package reflectpath

import (
	"regexp"
)

// KeyMatcher is a function that determines whether a property name
// matches the current path value.
type KeyMatcher func(key, pathname string) bool

// Resolver exposes methods to resolve a string path
// representation from a host structure and retrieve
// the matching value or type.
type Resolver struct {
	// KeyMatcher determines whether a structure key matches the current
	// path value. If not set, regular string comparison is used.
	KeyMatcher KeyMatcher
	// AllowedPatterns is a slice of regexp patterns representing
	// paths that are accessible by the Resolver.
	// If not set, there is no restriction.
	AllowedPatterns []string
}

func (r Resolver) safeMatchFunc(pathname string) func(string) bool {
	if r.KeyMatcher != nil {
		return func(key string) bool {
			return r.KeyMatcher(key, pathname)
		}
	}
	return func(key string) bool {
		return key == pathname
	}
}

func (r Resolver) isPathAllowed(pathRepr string) bool {
	for _, pattern := range r.AllowedPatterns {
		rgx := regexp.MustCompile(pattern)
		if rgx.MatchString(pathRepr) {
			return true
		}
	}
	return len(r.AllowedPatterns) == 0
}
