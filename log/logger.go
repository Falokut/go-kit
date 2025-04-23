package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type Encoder interface {
	Encode(field ...Field) ([]byte, error)
}

type exitFunc func(int)

type Adapter struct {
	out               []io.Writer
	encoder           Encoder
	enableTimestamp   bool
	deduplicateFields bool
	timeNow           func() time.Time

	level        *atomic.Uint32
	reportCaller bool
	exitFunc     exitFunc
}

func New(options ...Option) *Adapter {
	a := Default()
	for _, opt := range options {
		opt(a)
	}
	return a
}

func Default() *Adapter {
	level := &atomic.Uint32{}
	level.Store(uint32(DebugLevel))
	return &Adapter{
		out:             []io.Writer{os.Stdout},
		level:           level,
		exitFunc:        os.Exit,
		encoder:         JsonEncoder{},
		enableTimestamp: true,
		timeNow:         time.Now,
	}
}

func (l *Adapter) Close() error {
	return nil
}

func (l *Adapter) SetLogLevel(lvl Level) {
	l.level.Store(uint32(lvl))
}

func (l *Adapter) Level() Level {
	return Level(l.level.Load())
}

func (l *Adapter) IsLevelEnabled(level Level) bool {
	return Level(l.level.Load()) >= level
}

// Log will log a message at the level given as parameter.
func (l *Adapter) Log(ctx context.Context, level Level, msg any, fields ...Field) {
	if !l.IsLevelEnabled(level) {
		return
	}

	defaultLen := l.defaultFieldsLen(level)
	ctxFields := ContextLogValues(ctx)
	totalLen := defaultLen + len(ctxFields) + len(fields)

	toWrite := make([]Field, 0, totalLen)
	toWrite = l.appendDefaultFields(toWrite, level, msg)
	toWrite = append(toWrite, ctxFields...)
	toWrite = append(toWrite, fields...)

	if l.deduplicateFields {
		toWrite = deduplicateFields(toWrite)
	}
	l.write(toWrite...)
}

func (l *Adapter) Trace(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, TraceLevel, msg, fields...)
}

func (l *Adapter) Debug(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, DebugLevel, msg, fields...)
}

func (l *Adapter) Print(ctx context.Context, msg any, fields ...Field) {
	l.Info(ctx, msg, fields...)
}

func (l *Adapter) Info(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, InfoLevel, msg, fields...)
}

func (l *Adapter) Warn(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, WarnLevel, msg, fields...)
}

func (l *Adapter) Error(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, ErrorLevel, msg, fields...)
}

func (l *Adapter) Fatal(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, FatalLevel, msg, fields...)
	l.exitFunc(1)
}

func (l *Adapter) Panic(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, PanicLevel, msg, fields...)
}

func (l *Adapter) defaultFieldsLen(level Level) int {
	n := 2 // level + msg
	if l.enableTimestamp {
		n++
	}
	if l.reportCaller && level <= ErrorLevel {
		n += 2 // func + file
	}
	return n
}

func (l *Adapter) appendDefaultFields(dst []Field, level Level, msg any) []Field {
	if l.enableTimestamp {
		dst = append(dst, Time(FieldKeyTime, l.timeNow()))
	}
	dst = append(dst, String(FieldKeyLevel, level.String()))
	dst = append(dst, Any(FieldKeyMsg, msg))

	if l.reportCaller && level <= ErrorLevel {
		caller := getCaller()
		if caller.Function != "" {
			dst = append(dst, String(FieldKeyFunc, caller.Function))
		}
		if caller.File != "" {
			buf := strconv.AppendInt([]byte(caller.File+":"), int64(caller.Line), 10) // nolint:mnd
			dst = append(dst, String(FieldKeyFile, string(buf)))
		}
	}
	return dst
}

func (l *Adapter) write(fields ...Field) {
	serialized, err := l.encoder.Encode(fields...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode fields, %v\n", err)
		return
	}
	serialized = append(serialized, '\n')

	for _, out := range l.out {
		_, err := out.Write(serialized)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
		}
	}
}
