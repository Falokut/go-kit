package config

import "fmt"

type Listen struct {
	Host string `yaml:"host" env:"LISTEN_HOST" env-default:"0.0.0.0"`
	Port uint16 `yaml:"port" env:"LISTEN_PORT" env-default:"8080"`
}

func (l Listen) GetAddress() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}
