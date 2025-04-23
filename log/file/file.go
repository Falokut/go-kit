package file

import (
	"io"

	"github.com/natefinch/lumberjack"
)

type Config struct {
	Filepath   string
	MaxSizeMb  int
	MaxBackups int
	Compress   bool
}

func NewFileOutput(cfg Config) io.WriteCloser {
	return &lumberjack.Logger{
		Filename:   cfg.Filepath,
		MaxSize:    cfg.MaxSizeMb,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}
}
