// nolint
package maps

import (
	"fmt"
	"reflect"
)

func Flatten(value any, opts ...option) map[string]any {
	options := &flattenExpandOptions{prefix: "", sep: "."}
	for _, opt := range opts {
		opt(options)
	}
	result := make(map[string]any, 5)
	flatten(value, options, result)
	return result
}

func flatten(value any, opts *flattenExpandOptions, out map[string]any) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = reflect.Indirect(v)
	}

	if !v.IsValid() {
		if opts.prefix != "" {
			out[opts.prefix] = nil
		}
		return
	}

	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return
		}
		for _, key := range v.MapKeys() {
			childPrefix := joinKey(opts.prefix, key.String(), opts.sep)
			flatten(v.MapIndex(key).Interface(), &flattenExpandOptions{prefix: childPrefix, sep: opts.sep}, out)
		}

	case reflect.Struct:
		t := v.Type()
		for i := range v.NumField() {
			field := t.Field(i)
			if field.PkgPath != "" { // unexported
				continue
			}
			childKey := field.Name
			if field.Anonymous {
				childKey = ""
			}
			childPrefix := joinKey(opts.prefix, childKey, opts.sep)
			flatten(v.Field(i).Interface(), &flattenExpandOptions{prefix: childPrefix, sep: opts.sep}, out)
		}

	case reflect.Array, reflect.Slice:
		for i := range v.Len() {
			indexKey := fmt.Sprintf("[%d]", i)
			childPrefix := joinKey(opts.prefix, indexKey, opts.sep)
			flatten(v.Index(i).Interface(), &flattenExpandOptions{prefix: childPrefix, sep: opts.sep}, out)
		}

	default:
		if opts.prefix != "" {
			out[opts.prefix] = v.Interface()
		}
	}
}

func joinKey(prefix, key, sep string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + sep + key
}
