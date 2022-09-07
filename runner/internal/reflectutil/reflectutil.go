package reflectutil

import (
	"reflect"
	"strconv"
	"strings"
)

// KeyMatcher is a function that determines whether a property name
// matches the current path value.
type KeyMatcher func(key, pathname string) bool

// PathResolver exposes a method ResolvePath that resolves a path
// representation from a host structure.
type PathResolver struct {
	// KeyMatcher determines whether a structure key matches the current
	// path value. If not set, regular string comparison is used.
	KeyMatcher KeyMatcher
}

// ResolvePath resolves pathRepr starting from host and returns
// the corresponding value.
func (r PathResolver) ResolvePath(host interface{}, pathRepr string) reflect.Value {
	hostValue := reflect.ValueOf(host)
	pathStack := strings.Split(pathRepr, ".")
	return r.resolveRecursive(hostValue, pathStack)
}

func (r PathResolver) resolveRecursive(
	current reflect.Value,
	pathStack []string,
) reflect.Value {
	if len(pathStack) == 0 {
		return current
	}
	next := r.resolveProperty(current, pathStack[0])
	if len(pathStack) == 1 {
		return next
	}
	return r.resolveRecursive(next, pathStack[1:])
}

func (r PathResolver) resolveProperty(host reflect.Value, name string) reflect.Value {
	match := r.safeMatchFunc(name)
	kind := host.Kind()
	switch kind {
	case reflect.Struct:
		return propertyByNameFunc(host, match)
	case reflect.Map:
		return mapIndexFunc(host, match)
	case reflect.Slice:
		return sliceIndex(host, name)
	}
	panic(fmt.Sprintf("unhandled kind: %s", kind))
}

func (r PathResolver) safeMatchFunc(pathname string) func(string) bool {
	if r.KeyMatcher != nil {
		return func(key string) bool {
			return r.KeyMatcher(key, pathname)
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
	keyKind := host.Type().Key().Kind()
	iter := host.MapRange()
	for iter.Next() {
		key := iter.Key()
		switch keyKind {
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
	panic(fmt.Sprintf("unhandled key kind: %s", keyKind))
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
