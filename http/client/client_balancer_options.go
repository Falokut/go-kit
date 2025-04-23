package client

import (
	"net/http"
)

type clientBalancerOptions struct {
	cli        *http.Client
	clientOpts []Option
	schema     string
}

type ClientBalancerOption func(c *clientBalancerOptions)

// WithClientOptions appends passed opts to clientOpts
func WithClientOptions(opts ...Option) ClientBalancerOption {
	return func(c *clientBalancerOptions) {
		c.clientOpts = append(c.clientOpts, opts...)
	}
}

func WithHttpsSchema() ClientBalancerOption {
	return func(c *clientBalancerOptions) {
		c.schema = httpsSchema
	}
}

func WithClient(cli *http.Client) ClientBalancerOption {
	return func(c *clientBalancerOptions) {
		c.cli = cli
	}
}
