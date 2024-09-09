package log

import (
	"errors"
	"io"
	stdlog "log"
	"os"
)

type Config struct {
	Loglevel  Level
	Formatter FormatterConfig
	Output    OutputConfig
}

type OutputConfig struct {
	Filepath string
	Console  bool
}

const FormatterTypeJson = "json"

type FormatterConfig struct {
	Type string
	Json *JSONFormatterConfig
}

type JSONFormatterConfig struct {
	TimestampFormat   string
	DisableTimestamp  bool
	DisableHTMLEscape bool
	PrettyPrint       bool
	DataKey           string
}

func DefaultLoggerWithLevel(lvl Level) *Adapter {
	formatter := &JSONFormatter{}
	logger := &Adapter{
		Out:       os.Stdout,
		level:     lvl,
		Formatter: formatter,
		ExitFunc:  os.Exit,
	}
	return logger
}

func StdLoggerWithLevel(adapter Logger, level Level, withFields ...Field) *stdlog.Logger {
	logFunc, err := levelToFunc(adapter, level)
	if err != nil {
		panic(err)
	}
	stdLogger := stdlog.New(&LoggerWriter{logFunc: logFunc}, "", 0)
	return stdLogger
}

func NewFromConfig(cfg Config) (*Adapter, error) {
	if cfg.Loglevel > TraceLevel {
		return nil, errors.New("invalid log level")
	}

	formatter, err := getFormatter(cfg.Formatter)
	if err != nil {
		return nil, err
	}

	out, err := getOutputWriter(cfg.Output)
	if err != nil {
		return nil, err
	}
	logger := &Adapter{
		Out:       out,
		level:     cfg.Loglevel,
		Formatter: formatter,
		ExitFunc:  os.Exit,
	}
	return logger, nil
}

func getOutputWriter(cfg OutputConfig) (io.Writer, error) {
	var out io.Writer
	if cfg.Filepath != "" {
		file, err := os.OpenFile(cfg.Filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			return nil, err
		}
		out = file
	}
	if cfg.Console {
		if out != nil {
			out = io.MultiWriter(out, os.Stdout)
		} else {
			out = os.Stdout
		}
	}
	if out == nil {
		return nil, errors.New("invalid output")
	}
	return out, nil
}

func getFormatter(cfg FormatterConfig) (Formatter, error) {
	switch cfg.Type {
	case FormatterTypeJson:
		if cfg.Json == nil {
			return nil, errors.New("empty json formatter")
		}
		return &JSONFormatter{
			TimestampFormat:   cfg.Json.TimestampFormat,
			DisableTimestamp:  cfg.Json.DisableTimestamp,
			DisableHTMLEscape: cfg.Json.DisableHTMLEscape,
			PrettyPrint:       cfg.Json.PrettyPrint,
			DataKey:           cfg.Json.DataKey,
		}, nil
	default:
		return &JSONFormatter{}, nil
	}
}
