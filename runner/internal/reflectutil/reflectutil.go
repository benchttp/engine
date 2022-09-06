package reflectutil

import (
	"reflect"
	"strings"
)

// MatchFunc is a function that determines whether a property name
// matches the current path value.
type MatchFunc func(key, pathname string) bool

// ResolvePath resolves pathRepr starting from host and returns
// the corresponding value.
// It uses string comparison to determine whether the current path value
// matches a property name.
func ResolvePath(host interface{}, pathRepr string) reflect.Value {
	return resolvePath(host, pathRepr, nil)
}

// ResolvePath resolves pathRepr starting from host and returns
// the corresponding value.
// It uses the result of matchFunc to determine whether the current path value
// matches a property name.
func ResolvePathFunc(host interface{}, pathRepr string, matchFunc MatchFunc) reflect.Value {
	return resolvePath(host, pathRepr, matchFunc)
}

func resolvePath(
	host interface{},
	pathRepr string,
	matchFunc MatchFunc,
) reflect.Value {
	hostValue := reflect.ValueOf(host)
	pathStack := strings.Split(pathRepr, ".")
	return resolveRecursive(hostValue, pathStack, matchFunc)
}

func resolveRecursive(
	current reflect.Value,
	pathStack []string,
	matchFunc MatchFunc,
) reflect.Value {
	if len(pathStack) == 0 {
		return current
	}
	next := resolveProperty(current, pathStack[0], matchFunc)
	if len(pathStack) == 1 {
		return next
	}
	return resolveRecursive(next, pathStack[1:], matchFunc)
}

func resolveProperty(host reflect.Value, name string, matchFunc MatchFunc) reflect.Value {
	match := safeMatchFunc(matchFunc, name)
	switch host.Kind() {
	case reflect.Struct:
		if fieldMatch := host.FieldByNameFunc(match); fieldMatch.IsValid() {
			return fieldMatch
		}
	}
	return reflect.Value{}
}

func safeMatchFunc(matchFunc MatchFunc, pathname string) func(string) bool {
	if matchFunc != nil {
		return func(key string) bool {
			return matchFunc(key, pathname)
		}
	}
	return func(key string) bool {
		return key == pathname
	}
}
