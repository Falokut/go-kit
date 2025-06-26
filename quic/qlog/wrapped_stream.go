package qlog

import (
	"io"

	"github.com/Falokut/go-kit/quic"
)

type wrappedStream struct {
	quic.Stream
	reader io.Reader
	writer io.Writer
}

func (s *wrappedStream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *wrappedStream) Write(p []byte) (int, error) {
	return s.writer.Write(p)
}
