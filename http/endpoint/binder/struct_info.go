package binder

import (
	"reflect"
	"sync"
)

type fieldInfo struct {
	index     int
	fieldName string
	fieldType reflect.Type
	isSlice   bool
	elemKind  reflect.Kind
	isPtr     bool
	anonymous bool
}

type structInfo struct {
	fields []fieldInfo
}

type structCacheKey struct {
	typ reflect.Type
	tag string
}

var (
	structCacheMu sync.RWMutex
	structCache   = make(map[structCacheKey]*structInfo)
)

func getStructInfo(t reflect.Type, tag string) *structInfo {
	key := structCacheKey{typ: t, tag: tag}

	structCacheMu.RLock()
	if info, ok := structCache[key]; ok {
		structCacheMu.RUnlock()
		return info
	}
	structCacheMu.RUnlock()

	structCacheMu.Lock()
	defer structCacheMu.Unlock()

	info, ok := structCache[key]
	if ok {
		return info
	}

	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			continue
		}

		fieldName := getFieldName(field, tag)
		if fieldName == SkipParamFieldName {
			continue
		}

		fi := fieldInfo{
			index:     i,
			fieldName: fieldName,
			fieldType: field.Type,
			isSlice:   field.Type.Kind() == reflect.Slice,
			isPtr:     field.Type.Kind() == reflect.Ptr,
			anonymous: field.Anonymous,
		}

		if fi.isSlice {
			fi.elemKind = field.Type.Elem().Kind()
		} else if fi.isPtr && field.Type.Elem().Kind() == reflect.Slice {
			fi.elemKind = field.Type.Elem().Elem().Kind()
		}

		fields = append(fields, fi)
	}

	info = &structInfo{fields: fields}
	structCache[key] = info
	return info
}
