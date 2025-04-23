// nolint:exhaustive
package maps

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	arrayIndexRegexp = regexp.MustCompile(`\[\d*]`)
)

func Expand(flatMap map[string]any, opts ...option) any {
	options := &flattenExpandOptions{sep: "."}
	for _, opt := range opts {
		opt(options)
	}
	var dst any
	for path, value := range flatMap {
		parts := strings.Split(path, options.sep)
		dst = put(dst, parts, value)
	}
	return dst
}

// nolint:mnd
func put(dst any, path []string, value any) any {
	if len(path) == 0 {
		return value
	}

	p := path[0]
	index, isArray := getArrayIndex(p)
	// nolint:nestif
	if !isArray {
		if dst == nil {
			dst = make(map[string]any, 3)
		}
		m, ok := dst.(map[string]any)
		if ok {
			val, ok := m[p]
			if ok {
				m[p] = put(val, path[1:], value)
			} else {
				m[p] = put(nil, path[1:], value)
			}
		}
		return dst
	}

	if dst == nil {
		dst = make([]any, 0, 3)
	}

	arr, ok := dst.([]any)
	if !ok {
		return dst
	}

	i := len(arr)
	switch {
	case i == index:
		arr = append(arr, put(nil, path[1:], value))
	case i < index:
		toInsert := make([]any, 0, index-i)
		newItem := put(nil, path[1:], value)
		switch newItem.(type) {
		case []any:
			for i := range toInsert {
				toInsert[i] = make([]any, 0)
			}
		case map[string]any:
			for i := range toInsert {
				toInsert[i] = make(map[string]any, 0)
			}
		}
		arr = append(arr[:i], append(toInsert, newItem)...)
	default:
		arr[index] = put(arr[index], path[1:], value)
	}
	return arr
}

func getArrayIndex(part string) (int, bool) {
	index := arrayIndexRegexp.FindString(part)
	if index == "" {
		return 0, false
	}

	i, err := strconv.Atoi(index[1 : len(index)-1])
	if err != nil {
		return 0, false
	}

	return i, true
}
