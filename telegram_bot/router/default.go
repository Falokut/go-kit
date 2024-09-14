package router

import (
	"context"
	"fmt"
	"github.com/Falokut/go-kit/log"

	"github.com/Falokut/go-kit/telegram_bot"
)

func DefaultNotFoundHandler(_ context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error) {
	fromChat := msg.FromChat()
	if fromChat == nil {
		return nil, nil
	}
	return telegram_bot.NewMessage(fromChat.Id,
		fmt.Sprintf("для типа сообщения '%s' команда '%s' не найдена",
			msg.UpdateType(),
			msg.GetCommand(),
		),
	), nil
}

func DefaultMiddlewares(logger log.Logger) []Middleware {
	return []Middleware{
		RequestId(),
		Log(logger),
		ErrorHandler(logger),
		Recovery(),
	}
}
