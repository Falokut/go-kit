package router

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/requestid"
	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx/apierrors"
)

type BotError interface {
	BotResponse(chatId int64) tg_bot.Chattable
}

func ErrorHandler(logger log.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg tg_bot.Update) (tg_bot.Chattable, error) {
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

			// hide error details to prevent potential security leaks
			err = apierrors.NewInternalServiceError(err)
			return nil, err
		}
	}
}

func Log(logger log.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg tg_bot.Update) (tg_bot.Chattable, error) {
			fields := []log.Field{
				log.String("updateType", msg.UpdateType()),
			}
			sentFrom := msg.SentFrom()
			if sentFrom != nil {
				fields = append(fields,
					log.String("senderUsername", sentFrom.UserName),
					log.String("senderFirstName", sentFrom.FirstName),
					log.String("senderLastName", sentFrom.LastName),
				)
			}
			logger.Debug(ctx, "bot request", fields...)
			start := time.Now().UTC()
			resp, err := next(ctx, msg)
			stop := time.Now().UTC()
			logger.Debug(ctx,
				"bot response",
				log.Time("start", start),
				log.Time("stop", stop),
				log.Duration("durationMs", stop.Sub(start)),
			)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}
}

// nolint:nonamedreturns
func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg tg_bot.Update) (resp tg_bot.Chattable, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = fmt.Errorf("%v", r) // nolint:err113
				}
				stack := make([]byte, 4<<10) // nolint:mnd
				length := runtime.Stack(stack, false)
				err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			}()
			return next(ctx, msg)
		}
	}
}

func RequestId() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, msg tg_bot.Update) (tg_bot.Chattable, error) {
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
