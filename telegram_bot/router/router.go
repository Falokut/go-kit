package router

import (
	"context"
	"strings"

	"github.com/Falokut/go-kit/telegram_bot"
)

type node struct {
	handler HandlerFunc
	command string
}

type Router struct {
	trees           map[string][]*node
	notFoundHandler HandlerFunc
	middlewares     []Middleware
}

func NewRouter(middlewares ...Middleware) *Router {
	return &Router{
		trees:           make(map[string][]*node),
		middlewares:     middlewares,
		notFoundHandler: DefaultNotFoundHandler,
	}
}

func (r *Router) SetNotFoundHandler(notFoundHandler HandlerFunc) {
	r.notFoundHandler = notFoundHandler
}

func (r *Router) Handler(updateType string, command string, handler HandlerFunc) {
	n := &node{
		handler: handler,
		command: command,
	}
	root, ok := r.trees[updateType]
	if !ok {
		r.trees[updateType] = []*node{n}
		return
	}
	nodes := make([]*node, 0, len(root)+1)
	for _, node := range root {
		if node.command == command {
			continue
		}
		nodes = append(nodes, node)
	}
	nodes = append(nodes, n)
	r.trees[updateType] = nodes
}

func (r *Router) Handle(ctx context.Context, msg telegram_bot.Update) (telegram_bot.Chattable, error) {
	updateType := msg.UpdateType()
	root := r.trees[updateType]
	handler := r.notFoundHandler

	command := msg.GetCommand()
	for _, node := range root {
		if !strings.EqualFold(node.command, command) {
			continue
		}
		handler := node.handler
		for i := len(r.middlewares) - 1; i >= 0; i-- {
			handler = r.middlewares[i](handler)
		}
	}
	return handler(ctx, msg)
}
