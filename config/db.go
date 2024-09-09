package config

import (
	"fmt"
	"net/url"
)

type Database struct {
	Host        string `yaml:"host" env:"DATABASE_HOST" env-default:"localhost"`
	Port        uint32 `yaml:"port" env:"DATABASE_PORT" env-default:"5432"`
	Database    string `yaml:"database" env:"DATABASE_NAME" env-default:"postgres"`
	Username    string `yaml:"username" env:"DATABASE_USERNAME" env-default:"postgres"`
	Password    string `yaml:"password" env:"DATABASE_PASSWORD" env-default:"postgres"`
	Schema      string `yaml:"schema" env:"DATABASE_SCHEMA"`
	MaxOpenConn int    `yaml:"max_open_conn" env:"DATABASE_MAX_OPEN_CONN"`
}

func (c Database) ConnStr() string {
	u := url.URL{
		Scheme: "postgresql",
		User:   nil,
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Database,
	}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	query := url.Values{}
	if c.Schema != "" {
		query.Set("search_path", c.Schema)
	}
	u.RawQuery = query.Encode()
	return u.String()
}