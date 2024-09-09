package log

import "context"

type LogLevelSpecifier interface {
	LogLevel() Level
}

func LogLevelForError(err error) Level {
	logLevel := ErrorLevel
	specifier, ok := err.(LogLevelSpecifier)
	if ok {
		logLevel = specifier.LogLevel()
	}
	return logLevel
}

func LogLevelFuncForError(err error, logger Logger) func(ctx context.Context, message any, fields ...Field) {
	logLevel := LogLevelForError(err)
	switch logLevel {
	case ErrorLevel:
		return logger.Error
	case WarnLevel:
		return logger.Warn
	case InfoLevel:
		return logger.Info
	case DebugLevel:
		return logger.Debug
	default:
		return logger.Error
	}
}
