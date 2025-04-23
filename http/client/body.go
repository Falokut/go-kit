package client

import (
	"bytes"
	"github.com/Falokut/go-kit/json"
	"io"
	"net/http"
)

type RequestBodyWriter interface {
	Write(req *http.Request, w io.Writer) error
}

type jsonRequest struct {
	value any
}

func (j jsonRequest) Write(req *http.Request, w io.Writer) error {
	req.Header.Set("Content-Type", "application/json")
	return json.EncodeInto(w, j.value)
}

type plainRequest struct {
	value []byte
}

func (p plainRequest) Write(_ *http.Request, w io.Writer) error {
	_, err := w.Write(p.value)
	return err
}

type ResponseBodyReader interface {
	Read(r io.Reader) error
}

type jsonResponse struct {
	ptr any
}

func (j jsonResponse) Read(r io.Reader) error {
	buff, isBuff := r.(*bytes.Buffer)
	if isBuff {
		return json.Unmarshal(buff.Bytes(), j.ptr)
	}
	return json.NewDecoder(r).Decode(j.ptr)
}
