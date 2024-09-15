package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/router"
	"github.com/Falokut/go-kit/test"
)

type MockServer struct {
	wrapper endpoint.Wrapper
	srv     *httptest.Server
	router  *router.Router
}

func NewMock(t *test.Test) *MockServer {
	router := router.New()
	srv := httptest.NewServer(router)
	t.T().Cleanup(func() {
		srv.Close()
	})
	wrapper := endpoint.DefaultWrapper(t.Logger())
	return &MockServer{
		wrapper: wrapper,
		srv:     srv,
		router:  router,
	}
}

func (m *MockServer) Client() *http.Client {
	return m.srv.Client()
}

func (m *MockServer) BaseURL() string {
	return m.srv.URL
}

func (m *MockServer) POST(path string, handler any) *MockServer {
	return m.Mock(http.MethodPost, path, handler)
}

func (m *MockServer) GET(path string, handler any) *MockServer {
	return m.Mock(http.MethodGet, path, handler)
}

func (m *MockServer) Mock(method string, path string, handler any) *MockServer {
	m.router.Handler(method, path, m.wrapper.Endpoint(handler))
	return m
}
