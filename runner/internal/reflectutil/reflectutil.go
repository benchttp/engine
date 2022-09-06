package reflectutil

import (
	"reflect"
	"strings"
)

func ResolvePath(host interface{}, pathRepr string) reflect.Value {
	hostValue := reflect.ValueOf(host)
	pathStack := strings.Split(pathRepr, ".")
	return resolveRecursive(hostValue, pathStack)
}

func resolveRecursive(current reflect.Value, pathStack []string) reflect.Value {
	if len(pathStack) == 0 {
		return current
	}
	next := resolveProperty(current, pathStack[0])
	if len(pathStack) == 1 {
		return next
	}
	return resolveRecursive(next, pathStack[1:])
}

func resolveProperty(host reflect.Value, name string) reflect.Value {
	switch host.Kind() {
	case reflect.Struct:
		if fieldMatch := host.FieldByName(name); fieldMatch.IsValid() {
			return fieldMatch
		}
	}
	return reflect.Value{}
}
