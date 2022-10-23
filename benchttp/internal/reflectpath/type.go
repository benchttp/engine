package reflectpath

import (
	"fmt"
	"reflect"
	"strings"
)

// ResolveType resolves pathRepr starting from host and returns
// the matching type.
func (r Resolver) ResolveType(host interface{}, pathRepr string) reflect.Type {
	if !r.isPathAllowed(pathRepr) {
		return nil
	}
	hostValue := reflect.TypeOf(host)
	pathStack := strings.Split(pathRepr, ".")
	return r.resolveTypeRecursive(hostValue, pathStack)
}

func (r Resolver) resolveTypeRecursive(
	current reflect.Type,
	pathStack []string,
) reflect.Type {
	if len(pathStack) == 0 {
		return current
	}
	next := r.resolvePropertyType(current, pathStack[0])
	if len(pathStack) == 1 || next == nil {
		return next
	}
	return r.resolveTypeRecursive(next, pathStack[1:])
}

func (r Resolver) resolvePropertyType(host reflect.Type, name string) reflect.Type {
	match := r.safeMatchFunc(name)
	kind := host.Kind()
	switch kind {
	case reflect.Struct:
		return propertyTypeByNameFunc(host, match)
	case reflect.Map, reflect.Slice:
		return host.Elem()
	}
	panic(fmt.Sprintf("unhandled kind: %s", kind))
}

func propertyTypeByNameFunc(host reflect.Type, match func(string) bool) reflect.Type {
	if field, ok := host.FieldByNameFunc(match); ok {
		return field.Type
	}
	if method, ok := methodTypeByNameFunc(host, match); ok {
		return method.Type.Out(0)
	}
	return nil
}

func methodTypeByNameFunc(host reflect.Type, match func(string) bool) (reflect.Method, bool) {
	n := host.NumMethod()
	for i := 0; i < n; i++ {
		methodType := host.Method(i)
		if methodType.IsExported() && match(methodType.Name) {
			return methodType, true
		}
	}
	return reflect.Method{}, false
}
