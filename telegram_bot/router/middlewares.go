package router

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/log"
	requestid "github.com/Falokut/go-kit/requestId"
	"github.com/Falokut/go-kit/telegram_bot"
	"github.com/Falokut/go-kit/telegram_bot/apierrors"
)

type BotError interface {
	BotResponse(chatId int64) telegram_bot.Chattable
}

func ErrorHandler(logger log.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error) {
			resp, err := next(ctx, msg)
			if err == nil {
				return resp, nil
			}

			logFunc := log.LogLevelFuncForError(err, logger)
			logFunc(ctx, err)

			var botErr BotError
			chat := msg.FromChat()
			if errors.As(err, &botErr) && chat != nil {
				return botErr.BotResponse(chat.Id), nil
			}

			//hide error details to prevent potential security leaks
			err = apierrors.NewInternalServiceError(err)
			return nil, err
		}
	}
}

func Log(logger log.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error) {
			fields := []log.Field{
				log.Any("updateType", msg.UpdateType()),
			}
			sentFrom := msg.SentFrom()
			if sentFrom != nil {
				fields = append(fields,
					log.Any("senderUsername", sentFrom.UserName),
					log.Any("senderFirstName", sentFrom.FirstName),
					log.Any("senderLastName", sentFrom.LastName),
				)
			}
			logger.Debug(ctx, "bot request", fields...)
			start := time.Now().UTC()
			resp, err := next(ctx, msg)
			stop := time.Now().UTC()
			logger.Debug(ctx,
				"bot response",
				log.Any("start", start),
				log.Any("stop", stop),
				log.Any("durationMs", stop.Sub(stop).Milliseconds()),
			)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}
}

func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg telegram_bot.Update) (resp telegram_bot.Chattable, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = fmt.Errorf("%v", recovered)
				}
				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, false)
				err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			}()

			return next(ctx, msg)
		}
	}
}

func RequestId() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error) {
			reqId := requestid.FromContext(ctx)
			if reqId == "" {
				reqId = requestid.Next()
				ctx = requestid.ToContext(ctx, reqId)
			}
			ctx = log.ToContext(ctx, log.Any("requestId", reqId))
			return next(ctx, msg)
		}
	}
}
