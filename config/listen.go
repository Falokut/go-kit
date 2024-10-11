package config

import "fmt"

type Listen struct {
	Host string `yaml:"host" env:"LISTEN_HOST" env-default:"0.0.0.0" validate:"ip"`
	Port int    `yaml:"port" env:"LISTEN_PORT" env-default:"8080" validate:"gte=1024,lte=65535"`
}

func (l Listen) GetAddress() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}
