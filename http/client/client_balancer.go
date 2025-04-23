package client

import (
	"context"
	"net/http"
	"strings"

	"github.com/Falokut/go-kit/lb"
	"github.com/pkg/errors"
)

const (
	httpsSchema = "https://"
	httpSchema  = "http://"
)

type ClientBalancer struct {
	*Client
	hostManager *lb.RoundRobin
	schema      string
}

func NewClientBalancer(initialHosts []string, opts ...ClientBalancerOption) *ClientBalancer {
	options := &clientBalancerOptions{
		cli:        nil,
		clientOpts: nil,
		schema:     httpSchema,
	}
	for _, opt := range opts {
		opt(options)
	}
	initialHosts = addSchemaToHosts(options.schema, initialHosts)

	return &ClientBalancer{
		Client:      httpClient(options.cli, options.clientOpts...),
		hostManager: lb.NewRoundRobin(initialHosts),
		schema:      options.schema,
	}
}

func (c *ClientBalancer) Post(method string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPost, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Get(method string) *RequestBuilder {
	return NewRequestBuilder(http.MethodGet, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Put(method string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPut, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Delete(method string) *RequestBuilder {
	return NewRequestBuilder(http.MethodDelete, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Patch(method string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPatch, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Execute(ctx context.Context, builder *RequestBuilder) (*Response, error) {
	if c.GlobalRequestConfig().BaseUrl != "" {
		return c.Client.execute(ctx, builder)
	}

	host, err := c.hostManager.Next()
	if err != nil {
		return nil, errors.WithMessage(err, "host manager next")
	}

	return c.Client.execute(ctx, builder.BaseUrl(host))
}

func (c *ClientBalancer) Upgrade(hosts []string) {
	hosts = addSchemaToHosts(c.schema, hosts)
	c.hostManager.Upgrade(hosts)
}

func addSchemaToHosts(schema string, hosts []string) []string {
	for i, host := range hosts {
		shouldAddSchema := !strings.HasPrefix(host, httpSchema) && !strings.HasPrefix(host, httpsSchema)
		if shouldAddSchema {
			hosts[i] = schema + host
		}
	}
	return hosts
}

func httpClient(cli *http.Client, opts ...Option) *Client {
	if cli == nil {
		return New(opts...)
	}
	return NewWithClient(cli, opts...)
}
