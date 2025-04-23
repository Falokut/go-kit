// nolint:forcetypeassert,gochecknoglobals,mnd
package log

import "sync"

var dedupMapPool = sync.Pool{
	New: func() any {
		return make(map[string]struct{}, 64)
	},
}

func deduplicateFields(fields []Field) []Field {
	if len(fields) <= 1 {
		return fields
	}

	seen := dedupMapPool.Get().(map[string]struct{})
	for k := range seen {
		delete(seen, k)
	}

	deduped := make([]Field, 0, len(fields))
	for i := len(fields) - 1; i >= 0; i-- {
		f := fields[i]
		key := f.Name
		_, ok := seen[key]
		if !ok {
			seen[key] = struct{}{}
			deduped = append(deduped, f)
		}
	}

	dedupMapPool.Put(seen)

	for i, j := 0, len(deduped)-1; i < j; i, j = i+1, j-1 {
		deduped[i], deduped[j] = deduped[j], deduped[i]
	}

	return deduped
}
