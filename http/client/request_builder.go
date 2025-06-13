package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	libHttp "github.com/Falokut/go-kit/http"
)

var (
	ErrEmptyRequestUrl = errors.New("request url is empty")
)

type RequestBuilder struct {
	method            string
	url               string
	baseUrl           string
	headers           map[string]string
	basicAuth         *BasicAuth
	cookies           []*http.Cookie
	requestBody       RequestBodyWriter
	responseBody      ResponseBodyReader
	queryParams       map[string]any
	timeout           time.Duration
	statusCodeToError bool

	execute func(ctx context.Context, req *RequestBuilder) (*Response, error)
}

func NewRequestBuilder(
	method string,
	url string,
	cfg *GlobalRequestConfig,
	execute func(ctx context.Context, req *RequestBuilder) (*Response, error),
) *RequestBuilder {
	builder := &RequestBuilder{
		method:  method,
		url:     url,
		execute: execute,
	}
	cfg.configure(builder)
	return builder
}

func (b *RequestBuilder) BaseUrl(baseUrl string) *RequestBuilder {
	b.baseUrl = baseUrl
	return b
}

func (b *RequestBuilder) Header(name string, value string) *RequestBuilder {
	if b.headers == nil {
		b.headers = map[string]string{}
	}
	b.headers[name] = value
	return b
}

func (b *RequestBuilder) BearerAuth(token string) *RequestBuilder {
	return b.Header(libHttp.AuthorizationHeader, fmt.Sprint(libHttp.BearerToken, " ", token))
}

func (b *RequestBuilder) BasicAuth(ba BasicAuth) *RequestBuilder {
	b.basicAuth = &ba
	return b
}

func (b *RequestBuilder) Cookie(cookie *http.Cookie) *RequestBuilder {
	b.cookies = append(b.cookies, cookie)
	return b
}

func (b *RequestBuilder) RequestBody(body []byte) *RequestBuilder {
	b.requestBody = plainRequest{value: body}
	return b
}

func (b *RequestBuilder) JsonRequestBody(value any) *RequestBuilder {
	b.requestBody = jsonRequest{value: value}
	return b
}

// JsonResponseBody
// If response status code between 200 and 299, unmarshal response body to responsePtr
func (b *RequestBuilder) JsonResponseBody(responsePtr any) *RequestBuilder {
	b.responseBody = jsonResponse{ptr: responsePtr}
	return b
}

func (b *RequestBuilder) QueryParams(queryParams map[string]any) *RequestBuilder {
	b.queryParams = queryParams
	return b
}

// StatusCodeToError
// If set and Response.IsSuccess is false, Do return ErrorResponse as error
func (b *RequestBuilder) StatusCodeToError() *RequestBuilder {
	b.statusCodeToError = true
	return b
}

// Timeout
// Set per request attempt timeout, default timeout 15 seconds
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.timeout = timeout
	return b
}

func (b *RequestBuilder) Do(ctx context.Context) (*Response, error) {
	resp, err := b.execute(ctx, b)
	b.execute = nil
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() && b.statusCodeToError {
		body, err := resp.BodyCopy()
		if err != nil {
			return nil, err
		}
		return nil, ErrorResponse{
			Url:        resp.Raw.Request.URL,
			StatusCode: resp.StatusCode(),
			Body:       body,
		}
	}
	return resp, err
}

func (b *RequestBuilder) DoWithoutResponse(ctx context.Context) error {
	resp, err := b.Do(ctx)
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}

func (b *RequestBuilder) DoAndReadBody(ctx context.Context) ([]byte, int, error) {
	resp, err := b.Do(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Close()

	body, err := resp.BodyCopy()
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode(), nil
}

func (b *RequestBuilder) newHttpRequest(ctx context.Context) (*http.Request, error) {
	finalUrl, err := b.GetRequestUrl()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, b.method, finalUrl, nil)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (b *RequestBuilder) GetRequestUrl() (string, error) {
	var finalUrl string

	switch {
	case b.baseUrl == "":
		finalUrl = b.url
	case b.url == "":
		finalUrl = b.baseUrl
	default:
		base, err := url.Parse(b.baseUrl)
		if err != nil {
			return "", err
		}
		rel, err := url.Parse(b.url)
		if err != nil {
			return "", err
		}
		finalUrl = base.ResolveReference(rel).String()
	}

	if finalUrl == "" {
		return "", ErrEmptyRequestUrl
	}

	return finalUrl, nil
}
