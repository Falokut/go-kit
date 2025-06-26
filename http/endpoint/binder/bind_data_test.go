package binder_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Falokut/go-kit/http/endpoint/binder"
)

func TestBindDataSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(BindDataSuite))
}

type BindDataSuite struct {
	suite.Suite
}

type Embedded struct {
	Inner int `form:"inner"`
}

type TestStruct struct {
	ID      int       `form:"id"`
	Name    string    `form:"name"`
	Values  []string  `form:"values"`
	PtrVal  *int      `form:"ptrVal"`
	Embed   Embedded  `form:"embed"`
	EmbedP  *Embedded `form:"embedP"`
	SkipMe  string    `form:"-"`
	private string
}

func (s *BindDataSuite) TestBasicBinding() {
	values := map[string][]string{
		"id":          {"123"},
		"name":        {"tester"},
		"values":      {"a", "b", "c"},
		"ptrVal":      {"99"},
		"embed.inner": {"77"},
	}

	var ts TestStruct
	err := binder.BindData(values, &ts, binder.FormTag)
	s.Require().NoError(err)

	s.Equal(123, ts.ID)
	s.Equal("tester", ts.Name)
	s.Equal([]string{"a", "b", "c"}, ts.Values)
	s.NotNil(ts.PtrVal)
	s.Equal(99, *ts.PtrVal)
	s.Equal(77, ts.Embed.Inner)
	s.Empty(ts.SkipMe)
	s.Empty(ts.private)
}

func (s *BindDataSuite) TestEmbeddedPointerInitialized() {
	values := map[string][]string{
		"embedP.inner": {"42"},
	}

	var ts TestStruct
	s.Nil(ts.EmbedP)

	err := binder.BindData(values, &ts, binder.FormTag)
	s.Require().NoError(err)

	s.NotNil(ts.EmbedP)
	s.Equal(42, ts.EmbedP.Inner)
}

func (s *BindDataSuite) TestSkipField() {
	values := map[string][]string{
		"skipMe": {"should not bind"},
	}

	ts := TestStruct{SkipMe: "preset"}
	err := binder.BindData(values, &ts, binder.FormTag)
	s.Require().NoError(err)
	s.Equal("preset", ts.SkipMe)
}

func (s *BindDataSuite) TestInvalidIntConversion() {
	values := map[string][]string{
		"id": {"notanint"},
	}

	var ts TestStruct
	err := binder.BindData(values, &ts, binder.FormTag)

	s.Require().Error(err)
	s.Contains(err.Error(), "unmarshal field")
}

func (s *BindDataSuite) TestUnexportedFieldsAreIgnored() {
	type S struct {
		Exported   int    `form:"exported"`
		unexported string `form:"unexported"`
	}
	values := map[string][]string{
		"exported":   {"5"},
		"unexported": {"ignored"},
	}

	var st S
	err := binder.BindData(values, &st, binder.FormTag)
	s.Require().NoError(err)

	s.Equal(5, st.Exported)
	s.Empty(st.unexported)
}
