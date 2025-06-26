package qlog

type Option func(cfg *logConfig)

func WithLogBody(logBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logBody
		cfg.logResponseBody = logBody
	}
}

func WithLogRequestBody(logRequestBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

func WithLogResponseBody(logResponseBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}
