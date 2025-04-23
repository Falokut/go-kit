package app

import (
	"github.com/Falokut/go-kit/log/file"
)

type EncoderType string

const (
	JsonEncoderType      EncoderType = "json"
	PlainTextEncoderType EncoderType = "plain-text"
)

type LogConfig struct {
	FileOutput        *file.Config
	EncoderType       EncoderType
	DeduplicateFields bool
}
