package router

import (
	"context"

	"github.com/Falokut/go-kit/telegram_bot"
)

type HandlerFunc func(ctx context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error)
type Middleware func(next HandlerFunc) HandlerFunc
