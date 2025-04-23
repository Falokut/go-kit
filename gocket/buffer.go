package gocket

import (
	"bytes"
	"sync"
)

// nolint:gochecknoglobals,mnd
var (
	bpool = sync.Pool{New: func() any {
		return bytes.NewBuffer(make([]byte, 1024))
	}}
)

func getBuffer() *bytes.Buffer {
	b := bpool.Get().(*bytes.Buffer) // nolint:forcetypeassert
	b.Reset()
	return b
}

func putInBuffer(buf *bytes.Buffer) {
	bpool.Put(buf)
}
