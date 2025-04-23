package log_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Falokut/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonEncoder_EncodeBasicFields(t *testing.T) {
	t.Parallel()
	encoder := log.JsonEncoder{}

	fields := []log.Field{
		log.String("msg", "hello"),
		log.Int("code", 200),
		log.Bool("ok", true),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := `{"msg": "hello", "code": 200, "ok": true}`
	assert.JSONEq(t, expectedLog, string(result))
}

func TestPlainTextEncoder_EncodeBasicFields(t *testing.T) {
	t.Parallel()
	encoder := log.PlainTextEncoder{}

	fields := []log.Field{
		log.String("msg", "start"),
		log.Int("attempt", 1),
		log.Bool("success", false),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := `msg="start" attempt=1 success=false`
	assert.Equal(t, expectedLog, string(result))
}

func TestJsonEncoder_EncodeTimeDuration(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	encoder := log.JsonEncoder{}

	fields := []log.Field{
		log.Time("ts", now),
		log.Duration("elapsed", 2*time.Second),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := fmt.Sprintf(`{"ts": "%s", "elapsed": "2s"}`, now.Format(time.RFC3339))
	assert.Equal(t, expectedLog, string(result))
}

func TestPlainTextEncoder_EncodeTimeDuration(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	encoder := log.PlainTextEncoder{}

	fields := []log.Field{
		log.Time("ts", now),
		log.Duration("elapsed", 1500*time.Millisecond),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := fmt.Sprintf(`ts="%s" elapsed="1.5s"`, now.Format(time.RFC3339))
	assert.Equal(t, expectedLog, string(result))
}

func TestJsonEncoder_ArrayFields(t *testing.T) {
	t.Parallel()
	encoder := log.JsonEncoder{}

	fields := []log.Field{
		log.Strings("tags", []string{"a", "b", "c"}),
		log.Ints("scores", []int{1, 2, 3}),
		log.Bools("flags", []bool{true, false}),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	assert.JSONEq(t, `{"tags": ["a", "b", "c"], "scores": [1, 2, 3], "flags": [true, false]}`, string(result))
}

func TestPlainTextEncoder_ArrayFields(t *testing.T) {
	t.Parallel()
	encoder := log.PlainTextEncoder{}

	fields := []log.Field{
		log.Strings("tags", []string{"alpha", "beta"}),
		log.Floats64("ratios", []float64{1.1, 2.2}),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := `tags=["alpha", "beta"] ratios=[1.1, 2.2]`
	assert.Equal(t, expectedLog, string(result))
}

type User struct {
	Name  string
	Age   int
	Admin bool
}

func TestJsonEncoder_StructReflect(t *testing.T) {
	t.Parallel()
	encoder := log.JsonEncoder{}

	user := User{Name: "Bob", Age: 30, Admin: true}
	fields := []log.Field{
		log.Any("user", user),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := `{"user": {"Name": "Bob", "Age": 30, "Admin": true}}`
	assert.Equal(t, expectedLog, string(result))
}

func TestPlainTextEncoder_StructReflect(t *testing.T) {
	t.Parallel()
	encoder := log.PlainTextEncoder{}

	user := User{Name: "Alice", Age: 25, Admin: false}
	fields := []log.Field{
		log.Any("user", user),
	}

	result, err := encoder.Encode(fields...)
	require.NoError(t, err)
	expectedLog := `user={Name="Alice", Age=25, Admin=false}`
	assert.Equal(t, expectedLog, string(result))
}
