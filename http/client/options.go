package client

type Option func(c *Client)

func WithMiddlewares(mws ...Middleware) Option {
	return func(c *Client) {
		c.mws = append(c.mws, mws...)
	}
}
