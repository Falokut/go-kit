package client

import (
	"crypto/tls"
	"time"
)

type Option func(*Client)

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *Client) {
		c.tlsConfig = tlsConfig
	}
}

func WithMiddlewares(mws ...Middleware) Option {
	return func(c *Client) {
		c.mws = append(c.mws, mws...)
	}
}

func WithReadBufferSize(size int) Option {
	return func(c *Client) {
		c.readBufferSize = size
	}
}

func WithDialTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.dialTimeout = timeout
	}
}
