package apierrors

import (
	"fmt"
	"net/http"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/telegram_bot"
)

const (
	ErrCodeInternal = 900
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
	return New(true, ErrCodeInternal, "internal service error", err).
		WithLogLevel(log.ErrorLevel)
}

func NewBusinessError(errorCode int, errorMessage string, err error) Error {
	return New(false, errorCode, errorMessage, err).
		WithLogLevel(log.WarnLevel)
}

func (e Error) BotError(chatId int64) telegram_bot.Chattable {
	if e.httpStatusCode == http.StatusInternalServerError {
		return nil
	}
	return telegram_bot.NewMessage(chatId, e.ErrorMessage)
}

func New(
	internal bool,
	errorCode int,
	errorMessage string,
	err error,
) Error {
	return Error{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		cause:        err,
		level:        log.ErrorLevel,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.ErrorCode, e.ErrorMessage, e.cause)
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
