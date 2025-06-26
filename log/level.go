// nolint:err113
package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
)

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

type Level uint32 // nolint:recvcheck

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
	case TraceLevel:
		return logger.Trace, nil
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

// MarshalText marshals the Level to text. Note that the text representation
// drops the -Level suffix (see example).
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

// Set sets the level for the flag.Value interface.
func (l *Level) Set(s string) error {
	return l.UnmarshalText([]byte(s))
}

func (l *Level) unmarshalText(text []byte) bool {
	lvl, err := ParseLogLevel(string(text))
	if err != nil {
		return false
	}
	*l = lvl
	return true
}
