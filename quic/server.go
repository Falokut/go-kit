package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"sync/atomic"

	"github.com/Falokut/go-kit/log"
	"github.com/quic-go/quic-go"
	quicLib "github.com/quic-go/quic-go"
)

type StreamHandler interface {
	HandleStream(ctx context.Context, stream Stream) error
}

type service struct {
	delegate atomic.Value
}

type StreamHandlerFunc func(ctx context.Context, stream Stream) error

func (f StreamHandlerFunc) HandleStream(ctx context.Context, s Stream) error {
	return f(ctx, s)
}

type Server struct {
	service    *service
	tlsConfig  *tls.Config
	quicConfig *quicLib.Config

	middlewares []Middleware
	listener    atomic.Pointer[quicLib.Listener]

	cancel context.CancelFunc
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		service: &service{},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func DefaultServer(logger log.Logger) *Server {
	return NewServer(
		WithMiddlewares(
			Recovery(),
			ErrorHandler(logger),
			MaxRequestBodySize(DefaultMaxRequestBodySize),
			DecodeRequestMiddleware(),
			RequestId(),
		),
	)
}

func (s *Server) Listener() *quic.Listener {
	return s.listener.Load()
}

func (s *Server) Upgrade(h StreamHandler) {
	handler := ChainMiddleware(h, s.middlewares...)
	s.service.delegate.Store(handler)
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	if s.tlsConfig == nil {
		return errors.New("TLS config must be set") // nolint:err113
	}
	ctx, s.cancel = context.WithCancel(ctx)

	listener, err := quicLib.ListenAddr(addr, s.tlsConfig, s.quicConfig)
	if err != nil {
		return err
	}
	s.listener.Store(listener)

	for {
		conn, err := listener.Accept(ctx)
		switch {
		case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
			return nil
		case err != nil:
			return err
		default:
			go s.service.handleConnection(ctx, conn)
		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	listener := s.Listener()
	if listener != nil {
		return listener.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *service) handleConnection(ctx context.Context, conn *quicLib.Conn) {
	defer func() {
		_ = conn.CloseWithError(0, "connection closed")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return
		}

		handler, ok := s.delegate.Load().(StreamHandler)
		if !ok || handler == nil {
			_ = stream.Close()
			continue
		}

		go func(s *quicLib.Stream) {
			defer s.Close()
			_ = handler.HandleStream(s.Context(), wrapStream(s))
		}(stream)
	}
}
