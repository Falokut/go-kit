package buffer

import (
	"bytes"
)

type RequestBody struct {
	body *bytes.Buffer
}

func NewRequestBody(body []byte) RequestBody {
	return RequestBody{
		body: bytes.NewBuffer(body),
	}
}

func (r RequestBody) Read(p []byte) (int, error) {
	return r.body.Read(p)
}

func (r RequestBody) Close() error {
	return nil
}
