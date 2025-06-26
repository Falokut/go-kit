package quic

import (
	"crypto/tls"

	quicLib "github.com/quic-go/quic-go"
)

type Option func(*Server)

func WithTLSConfig(tlsConf *tls.Config) Option {
	return func(s *Server) {
		s.tlsConfig = tlsConf
	}
}

func WithQUICConfig(quicConf *quicLib.Config) Option {
	return func(s *Server) {
		s.quicConfig = quicConf
	}
}

func WithMiddlewares(mws ...Middleware) Option {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, mws...)
	}
}
