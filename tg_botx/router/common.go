package router

import (
	"context"

	"github.com/Falokut/go-kit/tg_bot"
)

type HandlerFunc func(ctx context.Context, msg tg_bot.Update) (tg_bot.Chattable, error)
type Middleware func(next HandlerFunc) HandlerFunc
