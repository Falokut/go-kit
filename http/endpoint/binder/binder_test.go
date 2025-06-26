package binder_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Falokut/go-kit/http/router"
	"github.com/julienschmidt/httprouter"

	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/stretchr/testify/suite"
)

type stubValidator struct {
	ok      bool
	details map[string]string
}

func (v *stubValidator) Validate(value any) (bool, map[string]string) {
	return v.ok, v.details
}

func TestRequestBinderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RequestBinderSuite))
}

type RequestBinderSuite struct {
	suite.Suite
	binder *binder.RequestBinder
}

func (s *RequestBinderSuite) SetupTest() {
	s.binder = binder.NewRequestBinder(&stubValidator{ok: true})
}

type testStruct struct {
	Id   string `path:"id"`
	Name string `query:"name"`
	Age  int    `json:"age"`
}

func (s *RequestBinderSuite) Test_Bind_GET_Success() {
	req := httptest.NewRequest(http.MethodGet, "/users/123?name=John", nil)
	req = addPathParams(req, map[string]string{"id": "123"})
	req.URL.RawQuery = "name=John"

	jsonBody := `{"age": 30}`
	req.Body = io.NopCloser(strings.NewReader(jsonBody))

	val, err := s.binder.Bind(s.T().Context(),
		"application/json", req, reflect.TypeOf(testStruct{}))

	s.Require().NoError(err)
	s.Equal("123", val.FieldByName("Id").String())
	s.Equal("John", val.FieldByName("Name").String())
	s.Equal(30, int(val.FieldByName("Age").Int()))
}

func (s *RequestBinderSuite) Test_Bind_ValidationError() {
	s.binder = binder.NewRequestBinder(&stubValidator{
		ok:      false,
		details: map[string]string{"Name": "required"},
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"age":30,"name":""}`))
	req.Header.Set("Content-Type", "application/json")

	_, err := s.binder.Bind(s.T().Context(), "application/json", req, reflect.TypeOf(testStruct{}))

	s.Require().Error(err)
	apiErr := apierrors.Error{}
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusBadRequest, apiErr.HttpStatusCode())
	s.Contains(apiErr.Error(), "validation errors")
}

func addPathParams(r *http.Request, params map[string]string) *http.Request {
	p := make(router.Params, 0, len(params))
	for k, v := range params {
		p = append(p, httprouter.Param{Key: k, Value: v})
	}
	ctx := context.WithValue(r.Context(), httprouter.ParamsKey, p)
	return r.WithContext(ctx)
}
