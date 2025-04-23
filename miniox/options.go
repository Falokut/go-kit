package miniox

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Option func(opt *minio.Options)

func WithCredentials(creds *credentials.Credentials) Option {
	return func(opt *minio.Options) {
		opt.Creds = creds
	}
}

func WithSecure(secure bool) Option {
	return func(opt *minio.Options) {
		opt.Secure = secure
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(opt *minio.Options) {
		opt.Transport = transport
	}
}

func OptionsFromConfig(cfg Config) []Option {
	return []Option{
		WithCredentials(cfg.getCredentials()),
		WithSecure(cfg.Secure),
	}
}
