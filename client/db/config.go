package db

import (
	"fmt"
	"net/url"
)

type Config struct {
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	Schema      string
	MaxOpenConn int
}

func (c Config) ConnStr() string {
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
