package app

import (
	"github.com/Falokut/go-kit/config"
)

type Option func(c *Config)
type LoggerConfigSupplier func(cfg *config.Config) LogConfig

type Config struct {
	ConfigOptions        []config.Option
	LoggerConfigSupplier LoggerConfigSupplier
}

func DefaultConfig() *Config {
	return &Config{}
}

func WithConfigOptions(opts ...config.Option) Option {
	return func(c *Config) {
		c.ConfigOptions = append(c.ConfigOptions, opts...)
	}
}
