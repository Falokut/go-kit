package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/Falokut/go-kit/quic"
	quicLib "github.com/quic-go/quic-go"
)

type RoundTripper interface {
	RoundTrip(ctx context.Context, req *quic.Request) (*quic.Response, error)
}

type RoundTripperFunc func(ctx context.Context, req *quic.Request) (*quic.Response, error)

func (f RoundTripperFunc) RoundTrip(ctx context.Context, req *quic.Request) (*quic.Response, error) {
	return f(ctx, req)
}

type Middleware func(next RoundTripper) RoundTripper

type Client struct {
	roundTripper RoundTripper
	mws          []Middleware

	connections   map[string]*quicLib.Conn
	connectionsMu sync.Mutex

	tlsConfig      *tls.Config
	readBufferSize int
	dialTimeout    time.Duration
}

// nolint:gosec,mnd
func NewClient(opts ...Option) *Client {
	c := &Client{
		connections: make(map[string]*quicLib.Conn),
		tlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"h3"},
		},
		readBufferSize: 4096,
		dialTimeout:    5 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Оборачиваем roundTrip middleware'ами
	rt := RoundTripper(RoundTripperFunc(c.roundTrip))
	for i := len(c.mws) - 1; i >= 0; i-- {
		rt = c.mws[i](rt)
	}
	c.roundTripper = rt

	return c
}

func (c *Client) Request(addr string) *RequestBuilder {
	return NewRequestBuilder(addr, c.execute)
}

func (c *Client) execute(ctx context.Context, rb *RequestBuilder) (*quic.Response, error) {
	req := &quic.Request{
		Address: rb.address,
		Timeout: rb.timeout,
		Body:    rb.body,
		Headers: rb.headers,
	}
	return c.roundTripper.RoundTrip(ctx, req)
}

func (c *Client) roundTrip(ctx context.Context, req *quic.Request) (*quic.Response, error) {
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	conn, err := c.getConnection(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	if len(req.Body) > 0 {
		_, err = stream.Write(req.Body)
		if err != nil {
			return nil, err
		}
		err = stream.Close()
		if err != nil {
			return nil, err
		}
	}

	var respBuf bytes.Buffer
	buf := make([]byte, c.readBufferSize)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			respBuf.Write(buf[:n])
		}
		switch {
		case errors.Is(err, io.EOF):
			return &quic.Response{Body: respBuf.Bytes()}, nil
		case err != nil:
			return nil, err
		}
	}
}

func (c *Client) getConnection(ctx context.Context, addr string) (*quicLib.Conn, error) {
	c.connectionsMu.Lock()
	defer c.connectionsMu.Unlock()

	conn, ok := c.connections[addr]
	if ok {
		if conn.Context().Err() == nil {
			return conn, nil
		}
		_ = conn.CloseWithError(0, "reconnect")
		delete(c.connections, addr)
	}

	dialCtx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()

	conn, err := quicLib.DialAddr(dialCtx, addr, c.tlsConfig, nil)
	if err != nil {
		return nil, err
	}

	c.connections[addr] = conn
	return conn, nil
}
