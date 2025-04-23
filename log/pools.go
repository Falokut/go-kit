// nolint:gochecknoglobals,mnd
package log

import (
	"strings"
	"sync"
)

const kb = 1024

var (
	bpool = sync.Pool{
		New: func() any {
			buf := &strings.Builder{}
			buf.Grow(4 * kb)
			return buf
		},
	}
)
