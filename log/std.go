package log

import stdlog "log"

func StdLoggerWithLevel(adapter Logger, level Level, withFields ...Field) *stdlog.Logger {
	logFunc, err := levelToFunc(adapter, level)
	if err != nil {
		panic(err)
	}
	stdLogger := stdlog.New(&LoggerWriter{logFunc: logFunc}, "", 0)
	return stdLogger
}
