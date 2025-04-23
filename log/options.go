package log

import (
	"io"
	"time"
)

type Option func(*Adapter)

func WithOutput(w ...io.Writer) Option {
	return func(a *Adapter) {
		a.out = w
	}
}

func WithLevel(level Level) Option {
	return func(a *Adapter) {
		a.level.Store(uint32(level))
	}
}

func WithEncoder(encoder Encoder) Option {
	return func(a *Adapter) {
		a.encoder = encoder
	}
}

func WithExitFunction(f exitFunc) Option {
	return func(a *Adapter) {
		a.exitFunc = f
	}
}

func WithReportCaller(reportCaller bool) Option {
	return func(a *Adapter) {
		a.reportCaller = reportCaller
	}
}

func WithTimestamp(enableTimestamp bool) Option {
	return func(a *Adapter) {
		a.enableTimestamp = enableTimestamp
	}
}

func WithTimeNow(timeNow func() time.Time) Option {
	return func(a *Adapter) {
		a.timeNow = timeNow
	}
}

func WithFieldsDeduplication(deduplicateFields bool) Option {
	return func(a *Adapter) {
		a.deduplicateFields = deduplicateFields
	}
}
