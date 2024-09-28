package client

import (
	"time"
)

type GlobalRequestConfig struct {
	Timeout   time.Duration
	BaseUrl   string
	Headers   map[string]string
}

func NewGlobalRequestConfig() *GlobalRequestConfig {
	return &GlobalRequestConfig{
		Timeout: 15 * time.Second,
	}
}

func (c *GlobalRequestConfig) configure(req *RequestBuilder) {
	req.timeout = c.Timeout
	req.baseUrl = c.BaseUrl
	for name, value := range c.Headers {
		req.Header(name, value)
	}
}