package router

import (
	"context"
	"fmt"
	"github.com/Falokut/go-kit/log"

	"github.com/Falokut/go-kit/tg_bot"
)

// nolint:ireturn,nilnil
func NotFoundHandlerWithMessage(_ context.Context, msg tg_bot.Update) (tg_bot.Chattable, error) {
	fromChat := msg.FromChat()
	if fromChat == nil {
		return nil, nil
	}
	return tg_bot.NewMessage(fromChat.Id,
		fmt.Sprintf("для типа сообщения '%s' команда '%s' не найдена",
			msg.UpdateType(),
			msg.GetCommand(),
		),
	), nil
}

// nolint:ireturn,nilnil
func DefaultNotFoundHandler(_ context.Context, msg tg_bot.Update) (tg_bot.Chattable, error) {
	return nil, nil
}

func DefaultMiddlewares(logger log.Logger) []Middleware {
	return []Middleware{
		RequestId(),
		Log(logger),
		ErrorHandler(logger),
		Recovery(),
	}
}
