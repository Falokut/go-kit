package binder_test

import (
	"net/http"
	"testing"

	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/stretchr/testify/suite"
)

type BindPathSuite struct {
	suite.Suite
}

func TestBindPathSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(BindPathSuite))
}

func (s *BindPathSuite) TestSimpleStructBinding() {
	type User struct {
		ID   int    `path:"id"`
		Name string `path:"name"`
	}

	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodGet,
		"/",
		nil,
	)
	s.Require().NoError(err)

	req = addPathParams(req, map[string]string{
		"id":   "42",
		"name": "Alice",
	})

	var user User
	err = binder.BindPath(req, &user)
	s.Require().NoError(err)

	s.Equal(42, user.ID)
	s.Equal("Alice", user.Name)
}

func (s *BindPathSuite) TestEmbeddedStructBinding() {
	type Embedded struct {
		Code string `path:"code"`
	}
	type Item struct {
		ID       int       `path:"id"`
		Embedded *Embedded `path:"embedded"`
	}

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/", nil)
	s.Require().NoError(err)

	req = addPathParams(req, map[string]string{
		"id":            "10",
		"code":          "xyz",
		"embedded.code": "abc",
	})

	var item Item
	err = binder.BindPath(req, &item)
	s.Require().NoError(err)

	s.Equal(10, item.ID)
	s.NotNil(item.Embedded)
	s.Equal("abc", item.Embedded.Code)
}

func (s *BindPathSuite) TestMissingParamsIgnored() {
	type Data struct {
		A int `path:"a"`
		B int `path:"b"`
	}

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/", nil)
	s.Require().NoError(err)

	req = addPathParams(req, map[string]string{
		"a": "5",
		// "b" missing
	})

	var data Data
	err = binder.BindPath(req, &data)
	s.Require().NoError(err)

	s.Equal(5, data.A)
	s.Equal(0, data.B) // zero value, since b param missing
}

func (s *BindPathSuite) TestInvalidConversionReturnsError() {
	type Data struct {
		Num int `path:"num"`
	}

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/", nil)
	s.Require().NoError(err)

	req = addPathParams(req, map[string]string{
		"num": "notanint",
	})

	var data Data
	err = binder.BindPath(req, &data)
	s.Require().Error(err)
	s.Contains(err.Error(), "unmarshal field")
}
