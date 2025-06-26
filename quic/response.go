package quic

import (
	"fmt"

	"github.com/Falokut/go-kit/json"
)

type Response struct {
	Body    []byte            `json:"body"`
	Headers map[string]string `json:"headers"`
}

func DecodeResponse(stream Stream) (*Response, error) {
	var resp Response
	decoder := json.NewDecoder(stream)
	err := decoder.Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &resp, nil
}

func (r *Response) Write(stream Stream) error {
	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(r); err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	return stream.Close()
}
