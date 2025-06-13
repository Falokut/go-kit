package client_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Falokut/go-kit/http/client"
	"github.com/stretchr/testify/suite"
)

func TestRequestBuilder(t *testing.T) {
	t.Parallel()
	suite.Run(t, &RequestBuilderSuite{})
}

type RequestBuilderSuite struct {
	suite.Suite
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_HasBaseUrl() {
	cfg := &client.GlobalRequestConfig{
		BaseUrl: "https://domain:80",
	}
	builder := client.NewRequestBuilder(http.MethodGet, "/my_url/method", cfg, executeStub)

	actualUrl, err := builder.GetRequestUrl()
	s.NoError(err)

	const expectedUrl = "https://domain:80/my_url/method"
	s.Equal(expectedUrl, actualUrl)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_WithoutBaseUrl() {
	cfg := &client.GlobalRequestConfig{}
	builder := client.NewRequestBuilder(http.MethodGet, "https://domain:443/my_url/method", cfg, executeStub)

	actualUrl, err := builder.GetRequestUrl()
	s.NoError(err)

	const expectedUrl = "https://domain:443/my_url/method"
	s.Equal(expectedUrl, actualUrl)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_EmptyUrls_ReturnsError() {
	cfg := &client.GlobalRequestConfig{}
	builder := client.NewRequestBuilder(http.MethodGet, "", cfg, executeStub)

	url, err := builder.GetRequestUrl()
	s.Empty(url)
	s.Error(err)
	s.ErrorIs(err, client.ErrEmptyRequestUrl)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_InvalidBaseUrl_ReturnsError() {
	cfg := &client.GlobalRequestConfig{
		BaseUrl: "http://[invalid-url",
	}
	builder := client.NewRequestBuilder(http.MethodGet, "/path", cfg, executeStub)

	url, err := builder.GetRequestUrl()
	s.Empty(url)
	s.Error(err)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_InvalidUrl_ReturnsError() {
	cfg := &client.GlobalRequestConfig{
		BaseUrl: "https://domain.com",
	}
	builder := client.NewRequestBuilder(http.MethodGet, "http://[invalid-url", cfg, executeStub)

	url, err := builder.GetRequestUrl()
	s.Empty(url)
	s.Error(err)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_BaseUrlWithTrailingSlash() {
	cfg := &client.GlobalRequestConfig{
		BaseUrl: "https://domain.com/api/",
	}
	builder := client.NewRequestBuilder(http.MethodGet, "method", cfg, executeStub)

	url, err := builder.GetRequestUrl()
	s.NoError(err)
	s.Equal("https://domain.com/api/method", url)
}

func (s *RequestBuilderSuite) Test_GetRequestUrl_AbsoluteUrlIgnoresBaseUrl() {
	cfg := &client.GlobalRequestConfig{
		BaseUrl: "https://domain.com",
	}
	absUrl := "https://other.com/path"
	builder := client.NewRequestBuilder(http.MethodGet, absUrl, cfg, executeStub)

	url, err := builder.GetRequestUrl()
	s.NoError(err)
	s.Equal(absUrl, url)
}

func executeStub(ctx context.Context, req *client.RequestBuilder) (*client.Response, error) {
	return nil, nil
}
