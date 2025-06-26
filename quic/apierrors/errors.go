package apierrors

import (
	"fmt"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/quic"
)

type Error struct {
	errorCode    uint64
	errorMessage string
	details      map[string]any

	cause error
	level log.Level
}

func NewInternalServiceError(err error) Error {
	return New(quic.InternalErrorCode, "internal service error", err).
		WithLogLevel(log.ErrorLevel)
}

func NewBusinessError(errorCode uint64, errorMessage string, err error) Error {
	return New(errorCode, errorMessage, err).
		WithLogLevel(log.WarnLevel)
}

func New(
	errorCode uint64,
	errorMessage string,
	err error,
) Error {
	return Error{
		errorCode:    errorCode,
		errorMessage: errorMessage,
		cause:        err,
		level:        log.ErrorLevel,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.errorCode, e.errorMessage, e.cause)
}

func (e Error) ErrorCode() uint64 {
	return e.errorCode
}

func (e Error) WithLogLevel(level log.Level) Error {
	e.level = level
	return e
}

func (e Error) WithDetails(details map[string]any) Error {
	e.details = details
	return e
}

func (e Error) LogLevel() log.Level {
	return e.level
}
