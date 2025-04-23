package schema

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Id string

const EmptyId Id = ""

// nolint:err113
func (id Id) Validate() error {
	u, err := url.Parse(id.String())
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Hostname() == "" {
		return errors.New("missing hostname")
	}
	if !strings.Contains(u.Hostname(), ".") {
		return errors.New("hostname does not look valid")
	}
	if u.Path == "" {
		return errors.New("path is expected")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("unexpected schema")
	}
	return nil
}

func (id Id) Anchor(name string) Id {
	b := id.Base()
	return Id(b.String() + "#" + name)
}

func (id Id) Def(name string) Id {
	b := id.Base()
	return Id(b.String() + "#/$defs/" + name)
}

func (id Id) Add(path string) Id {
	b := id.Base()
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return Id(b.String() + path)
}

func (id Id) Base() Id {
	s := id.String()
	i := strings.LastIndex(s, "#")
	if i != -1 {
		s = s[0:i]
	}
	s = strings.TrimRight(s, "/")
	return Id(s)
}

func (id Id) String() string {
	return string(id)
}
