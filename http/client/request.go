package client

import (
	"net/http"
	"time"
)

type Request struct {
	Raw *http.Request

	body    []byte
	timeout time.Duration
}

func (r *Request) Body() []byte {
	return r.body
}
