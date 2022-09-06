package reflectutil

import (
	"reflect"
	"strconv"
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
		return propertyByNameFunc(host, match)
	case reflect.Map:
		return mapIndexFunc(host, match)
	case reflect.Slice:
		return sliceIndex(host, name)
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

func propertyByNameFunc(host reflect.Value, match func(string) bool) reflect.Value {
	if fieldMatch := host.FieldByNameFunc(match); fieldMatch.IsValid() {
		return fieldMatch
	}
	if methodMatch := methodByNameFunc(host, match); methodMatch.IsValid() {
		return methodMatch.Call([]reflect.Value{})[0]
	}
	return reflect.Value{}
}

func methodByNameFunc(host reflect.Value, match func(string) bool) reflect.Value {
	n := host.NumMethod()
	for i := 0; i < n; i++ {
		methodType := host.Type().Method(i)
		if methodType.IsExported() && match(methodType.Name) {
			return host.Method(i)
		}
	}
	return reflect.Value{}
}

func mapIndexFunc(host reflect.Value, match func(string) bool) reflect.Value {
	iter := host.MapRange()
	for iter.Next() {
		switch key := iter.Key(); key.Kind() {
		case reflect.String:
			if match(key.String()) {
				return iter.Value()
			}
		case reflect.Int:
			if match(strconv.Itoa(int(key.Int()))) {
				return iter.Value()
			}
		}
	}
	return reflect.Value{}
}

func sliceIndex(host reflect.Value, istr string) reflect.Value {
	i, err := strconv.Atoi(istr)
	if err != nil || i >= host.Len() {
		return reflect.Value{}
	}
	if elemMatch := host.Index(i); elemMatch.IsValid() {
		return elemMatch
	}
	return reflect.Value{}
}
