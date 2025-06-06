package fake_test

import (
	"testing"
	"time"

	"github.com/Falokut/go-kit/test/fake"
	"github.com/stretchr/testify/require"
)

type SomeStruct struct {
	A string
	B bool
}

func Test(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	intValue := fake.It[int]()
	require.NotEmpty(intValue)

	stringSlice := fake.It[[]string]()
	require.NotEmpty(stringSlice)

	structSlice := fake.It[[]SomeStruct]()
	t.Log(structSlice)
	require.NotEmpty(structSlice)

	time := fake.It[time.Time]()
	require.False(time.IsZero())
}
