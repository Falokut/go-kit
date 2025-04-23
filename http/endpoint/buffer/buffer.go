package buffer

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Buffer struct {
	http.ResponseWriter
	requestBuffer  *bytes.Buffer
	responseBuffer *bytes.Buffer
	statusCode     int
}

// nolint:mnd
func New() *Buffer {
	return &Buffer{
		requestBuffer:  bytes.NewBuffer(make([]byte, 1024)),
		responseBuffer: bytes.NewBuffer(make([]byte, 1024)),
	}
}

func (m *Buffer) Reset(w http.ResponseWriter) {
	m.ResponseWriter = w
	m.statusCode = 0
	m.responseBuffer.Reset()
	m.requestBuffer.Reset()
}

func (m *Buffer) Write(b []byte) (int, error) {
	n, err := m.ResponseWriter.Write(b)
	if err != nil {
		return n, errors.WithMessage(err, "write to response writer")
	}

	n, err = m.responseBuffer.Write(b)
	if err != nil {
		return 0, errors.WithMessage(err, "write to response buffer")
	}

	return n, nil
}

func (m *Buffer) WriteHeader(statusCode int) {
	m.statusCode = statusCode
	m.ResponseWriter.WriteHeader(statusCode)
}

func (m *Buffer) ResponseBody() []byte {
	return m.responseBuffer.Bytes()
}

func (m *Buffer) RequestBody() []byte {
	return m.requestBuffer.Bytes()
}

func (m *Buffer) ReadRequestBody(r io.Reader) error {
	_, err := io.Copy(m.requestBuffer, r)
	return err
}

func (m *Buffer) StatusCode() int {
	if m.statusCode == 0 {
		return http.StatusOK
	}
	return m.statusCode
}
