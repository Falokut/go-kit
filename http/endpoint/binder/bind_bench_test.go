package binder_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/Falokut/go-kit/http/endpoint/binder"
	"github.com/stretchr/testify/require"
)

func BenchmarkBindData_SimpleStruct(b *testing.B) {
	type benchStruct struct {
		ID     int      `form:"id"`
		Name   string   `form:"name"`
		Values []string `form:"values"`
		PtrVal *int     `form:"ptrVal"`
		SkipMe string   `form:"-"`
	}

	values := map[string][]string{
		"id":     {"123"},
		"name":   {"testname"},
		"values": {"a", "b", "c"},
		"ptrVal": {"42"},
	}

	var dest benchStruct

	b.ReportAllocs()
	for b.Loop() {
		err := binder.BindData(values, &dest, binder.FormTag)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBindPath_Item(b *testing.B) {
	type benchStruct struct {
		ID       int `path:"id"`
		Embedded *struct {
			Code string `path:"code"`
		} `path:"embedded"`
	}

	params := map[string]string{
		"id":            "123",
		"embedded.code": "abc123",
	}

	var dest benchStruct
	req, _ := http.NewRequest("GET", "/", nil)
	req = addPathParams(req, params)

	b.ReportAllocs()
	for b.Loop() {
		err := binder.BindPath(req, &dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRequestBinder_Bind(b *testing.B) {
	type BenchStruct struct {
		ID    string `path:"id"`
		Code  string `path:"code"`
		Query string
	}

	params := map[string]string{
		"id":   "123",
		"code": "abc123",
	}

	req, err := http.NewRequest("GET", "/?query=hello", nil)
	require.NoError(b, err)
	req = addPathParams(req, params)
	rb := binder.NewRequestBinder(&stubValidator{ok: true})
	benchStructType := reflect.TypeOf(BenchStruct{})

	b.ReportAllocs()
	for b.Loop() {
		_, err := rb.Bind(
			b.Context(),
			"",
			req,
			benchStructType,
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
