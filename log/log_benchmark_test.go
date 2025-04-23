// nolint:funlen
package log_test

import (
	"io"
	"testing"

	"github.com/Falokut/go-kit/log"
)

type nopEncoder struct{}

func (nopEncoder) Encode(...log.Field) ([]byte, error) {
	return nil, nil
}

type benchCase struct {
	name   string
	logger *log.Adapter
	fields []log.Field
}

func BenchmarkLogger(b *testing.B) {
	cases := generateLogTestCases()

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for b.Loop() {
				c.logger.Info(b.Context(), "The quick brown fox jumps over the lazy dog", c.fields...)
			}
		})
		b.Run(c.name+"_WithCtxFields", func(b *testing.B) {
			for b.Loop() {
				ctx := log.ToContext(b.Context(), log.String("key", "value"),
					log.Int("code", 42),
					log.Bool("isActive", true),
					log.Float32("price", 99.99),
					log.String("category", "test"),
					log.String("userID", "abc123"),
					log.Float64("housePrice", 9999.99),
					log.String("requestId", "uuid"),
				)
				c.logger.Info(ctx, "The quick brown fox jumps over the lazy dog", c.fields...)
			}
		})

		b.Run(c.name+"_WithReportCaller", func(b *testing.B) {
			for b.Loop() {
				c.logger.Error(b.Context(), "The quick brown fox jumps over the lazy dog", c.fields...)
			}
		})
	}
}

func newLogger(encoder log.Encoder) *log.Adapter {
	options := []log.Option{
		log.WithOutput(io.Discard),
		log.WithReportCaller(true),
	}

	if encoder != nil {
		options = append(options, log.WithEncoder(encoder))
	}

	return log.New(options...)
}

func newLoggerWithDeduplication(encoder log.Encoder) *log.Adapter {
	options := []log.Option{
		log.WithOutput(io.Discard),
		log.WithReportCaller(true),
		log.WithFieldsDeduplication(true),
	}

	if encoder != nil {
		options = append(options, log.WithEncoder(encoder))
	}

	return log.New(options...)
}

func generateLogTestCases() []benchCase {
	fields := []log.Field{
		log.String("key", "value"),
		log.Int("code", 42),
		log.Bool("isActive", true),
		log.Float32("price", 99.99),
		log.String("category", "test"),
		log.String("userID", "abc123"),
	}
	return []benchCase{
		{
			name:   "NopEncoder_NoFields",
			logger: newLogger(nopEncoder{}),
			fields: nil,
		},
		{
			name:   "NopEncoder_WithFields",
			logger: newLogger(nopEncoder{}),
			fields: fields,
		},
		{
			name:   "NopEncoder_WithFields_WithDeduplication",
			logger: newLoggerWithDeduplication(nopEncoder{}),
			fields: append(fields, fields...),
		},
		{
			name:   "JsonEncoder_NoFields",
			logger: newLogger(log.JsonEncoder{}),
			fields: nil,
		},
		{
			name:   "JsonEncoder_WithFields",
			logger: newLogger(log.JsonEncoder{}),
			fields: fields,
		},
		{
			name:   "JsonEncoder_WithFields_WithDeduplication",
			logger: newLoggerWithDeduplication(log.JsonEncoder{}),
			fields: append(fields, fields...),
		},
		{
			name:   "PlainTextEncoder_NoFields",
			logger: newLogger(log.PlainTextEncoder{}),
			fields: nil,
		},
		{
			name:   "PlainTextEncoder_WithFields",
			logger: newLogger(log.PlainTextEncoder{}),
			fields: fields,
		},
		{
			name:   "PlainTextEncoder_WithFields_WithDeduplication",
			logger: newLoggerWithDeduplication(log.PlainTextEncoder{}),
			fields: append(fields, fields...),
		},
	}
}
