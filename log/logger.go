package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/utils/maps"
)

type Logger interface {
	Log(ctx context.Context, level Level, msg any, fields ...Field)
	Trace(ctx context.Context, msg any, fields ...Field)
	Debug(ctx context.Context, msg any, fields ...Field)
	Info(ctx context.Context, msg any, fields ...Field)
	Warn(ctx context.Context, msg any, fields ...Field)
	Warning(ctx context.Context, msg any, fields ...Field)
	Error(ctx context.Context, msg any, fields ...Field)
	Fatal(ctx context.Context, msg any, fields ...Field)
	Panic(ctx context.Context, msg any, fields ...Field)
}

type Adapter struct {
	Out          io.Writer
	Formatter    Formatter
	ReportCaller bool
	level        Level
	mu           MutexWrap
	ExitFunc     exitFunc
}

type exitFunc func(int)

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}

func (l *Adapter) log(level Level, msg string, data Fields) {
	entry := &Entry{
		ReportCaller: l.ReportCaller,
		Data:         data,
		Time:         time.Now(),
		Level:        level,
		Caller:       getCaller(),
		Message:      msg,
	}
	l.write(entry)
}

func (l *Adapter) write(entry *Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	serialized, err := l.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		return
	}
	if _, err := l.Out.Write(serialized); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

func (l *Adapter) Level() Level {
	return Level(atomic.LoadUint32((*uint32)(&l.level)))
}

// IsLevelEnabled checks if the log level of the logger is greater than the level param
func (l *Adapter) IsLevelEnabled(level Level) bool {
	return l.Level() >= level
}

// SetFormatter sets the logger formatter.
func (logger *Adapter) SetFormatter(formatter Formatter) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Formatter = formatter
}

// SetOutput sets the logger output.
func (logger *Adapter) SetOutput(output io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Out = output
}

func (logger *Adapter) SetReportCaller(reportCaller bool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.ReportCaller = reportCaller
}

// Log will log a message at the level given as parameter.
// Warning: using Log at Panic or Fatal level will not respectively Panic nor Exit.
// For this behaviour logger.Panic or logger.Fatal should be used instead.
func (l *Adapter) Log(ctx context.Context, level Level, msg any, fields ...Field) {
	if l.IsLevelEnabled(level) {
		contextData := ContextLogValues(ctx)
		fieldData := toFieldsMap(fields...)
		data := maps.MergeMaps(contextData, fieldData)
		l.log(level, fmt.Sprint(msg), data)
	}
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

func (l *Adapter) Warning(ctx context.Context, msg any, fields ...Field) {
	l.Warn(ctx, msg, fields...)
}

func (l *Adapter) Error(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, ErrorLevel, msg, fields...)
}

func (l *Adapter) Fatal(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, FatalLevel, msg, fields...)
	l.ExitFunc(1)
}

func (l *Adapter) Panic(ctx context.Context, msg any, fields ...Field) {
	l.Log(ctx, PanicLevel, msg, fields...)
}
