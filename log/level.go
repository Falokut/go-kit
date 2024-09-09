package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Level uint32

// These are the different logging levels. You can set the logging level to log
// on your instance of logger
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARNING"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	case TraceLevel:
		return "TRACE"
	}
	return ""
}

func levelToFunc(logger Logger, lvl Level) (func(ctx context.Context, msg any, fields ...Field), error) {
	switch lvl {
	case DebugLevel:
		return logger.Debug, nil
	case InfoLevel:
		return logger.Info, nil
	case WarnLevel:
		return logger.Warn, nil
	case ErrorLevel:
		return logger.Error, nil
	case PanicLevel:
		return logger.Panic, nil
	case FatalLevel:
		return logger.Fatal, nil
	}
	return nil, fmt.Errorf("unrecognized level: %q", lvl)
}

func ParseLogLevel(lvl string) (Level, error) {
	switch {
	case strings.EqualFold(lvl, "PANIC"):
		return PanicLevel, nil
	case strings.EqualFold(lvl, "FATAL"):
		return FatalLevel, nil
	case strings.EqualFold(lvl, "ERROR"):
		return ErrorLevel, nil
	case strings.EqualFold(lvl, "WARNING") || strings.EqualFold(lvl, "WARN"):
		return WarnLevel, nil
	case strings.EqualFold(lvl, "INFO"):
		return InfoLevel, nil
	case strings.EqualFold(lvl, "DEBUG"):
		return DebugLevel, nil
	case strings.EqualFold(lvl, "TRACE"):
		return TraceLevel, nil
	}
	return Level(0), errors.New("unknown log level")
}
