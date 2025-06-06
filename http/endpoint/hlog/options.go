package hlog

type Option func(cfg *logConfig)

func WithContentTypes(logBodyContentTypes []string) Option {
	return func(cfg *logConfig) {
		cfg.logBodyContentTypes = logBodyContentTypes
	}
}

func WithLogBody(logBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

func WithLogResponseBody(logResponseBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

func WithLogRequestBody(logRequestBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}
