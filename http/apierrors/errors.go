package apierrors

import (
	"fmt"
	"net/http"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/log"
	"github.com/pkg/errors"
)

const (
	ErrCodeInvalidRange = 800
	ErrCodeInternal     = 900
	ErrCodeUnauthorized = 700
	ErrCodeForbidden    = 701
)

type Error struct {
	ErrorCode    int
	ErrorMessage string
	Details      map[string]any

	httpStatusCode int
	cause          error
	level          log.Level
}

func NewInternalServiceError(err error) Error {
	return New(http.StatusInternalServerError, ErrCodeInternal, "internal service error", err).
		WithLogLevel(log.ErrorLevel)
}

func NewBusinessError(errorCode int, errorMessage string, err error) Error {
	return New(http.StatusBadRequest, errorCode, errorMessage, err).
		WithLogLevel(log.WarnLevel)
}

func NewForbiddenError(errorMsg string) Error {
	return New(http.StatusForbidden, ErrCodeForbidden, errorMsg, errors.New(errorMsg))
}

func NewUnauthorizedError(errorMsg string) Error {
	return New(http.StatusUnauthorized, ErrCodeUnauthorized, errorMsg, errors.New(errorMsg))
}

func NewRangeUnacceptableError(errorMsg string) Error {
	return New(http.StatusRequestedRangeNotSatisfiable, ErrCodeInvalidRange, errorMsg, errors.New(errorMsg))
}

func New(
	httpStatusCode int,
	errorCode int,
	errorMessage string,
	err error,
) Error {
	return Error{
		ErrorCode:      errorCode,
		ErrorMessage:   errorMessage,
		httpStatusCode: httpStatusCode,
		cause:          err,
		level:          log.ErrorLevel,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.ErrorCode, e.ErrorMessage, e.cause)
}

func (e Error) HttpStatusCode() int {
	return e.httpStatusCode
}

func (e Error) WriteError(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.httpStatusCode)
	err := json.NewEncoder(w).Encode(e)
	if err != nil {
		return errors.WithMessage(err, "json encode error")
	}
	return nil
}

func (e Error) WithLogLevel(level log.Level) Error {
	e.level = level
	return e
}

func (e Error) WithDetails(details map[string]any) Error {
	e.Details = details
	return e
}

func (e Error) LogLevel() log.Level {
	return e.level
}
