package quic

import (
	"context"
	quicLib "github.com/quic-go/quic-go"
	"time"
)

type Stream interface {
	StreamID() int64
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	CancelWrite(errorCode uint64)
	CancelRead(errorCode uint64)
	Context() context.Context
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetDeadline(t time.Time) error
}

type streamWrapper struct {
	*quicLib.Stream
}

func wrapStream(s *quicLib.Stream) *streamWrapper {
	return &streamWrapper{Stream: s}
}

func (w *streamWrapper) StreamID() int64 {
	return w.StreamID()
}

func (w *streamWrapper) CancelWrite(errorCode uint64) {
	w.Stream.CancelWrite(quicLib.StreamErrorCode(errorCode))
}

func (w *streamWrapper) CancelRead(errorCode uint64) {
	w.Stream.CancelRead(quicLib.StreamErrorCode(errorCode))
}
