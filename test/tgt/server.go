package tgt

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/httpt"
	"github.com/Falokut/go-kit/tg_bot"
)

type BotServerMock struct {
	*httpt.MockServer
	updates           atomic.Value
	lastSendedMessage atomic.Value
}

func NewBotServerMock(test *test.Test, mock *httpt.MockServer, botInfo tg_bot.User) *BotServerMock {
	botServerMock := &BotServerMock{
		MockServer:        mock,
		updates:           atomic.Value{},
		lastSendedMessage: atomic.Value{},
	}
	botServerMock.POST("/test/getMe", func() tg_bot.ApiResponse {
		result, err := json.Marshal(botInfo)
		test.Assert().NoError(err)
		return tg_bot.ApiResponse{
			Ok:     true,
			Result: result,
		}
	})
	botServerMock.POST("/test/getUpdates", func(r *http.Request) tg_bot.ApiResponse {
		if r.URL.Query().Has("timeout") {
			timeoutVal := r.URL.Query().Get("timeout")
			timeout, err := strconv.Atoi(timeoutVal)
			test.Assert().NoError(err)
			time.Sleep(time.Duration(timeout) * time.Second)
		}
		updates := botServerMock.updates.Load().([]tg_bot.Update) // nolint:forcetypeassert
		offset := 0
		var err error
		if r.URL.Query().Has("offset") {
			offsetVal := r.URL.Query().Get("offset")
			offset, err = strconv.Atoi(offsetVal)
			test.Assert().NoError(err)
		}
		if offset > len(updates) {
			updates = []tg_bot.Update{}
		} else {
			updates = updates[offset:]
		}

		result, err := json.Marshal(updates)
		test.Assert().NoError(err)
		return tg_bot.ApiResponse{
			Ok:     true,
			Result: result,
		}
	})
	botServerMock.POST("/test/setMyCommands", func() tg_bot.ApiResponse {
		return tg_bot.ApiResponse{Ok: true}
	})
	botServerMock.POST("/test/deleteMyCommands", func() tg_bot.ApiResponse {
		return tg_bot.ApiResponse{Ok: true}
	})
	botServerMock.POST("/test/sendMessage", func(r *http.Request) tg_bot.ApiResponse {
		botServerMock.lastSendedMessage.Store(r.URL.Query())
		return tg_bot.ApiResponse{Ok: true}
	})
	return botServerMock
}

func (b *BotServerMock) SetUpdates(updates []tg_bot.Update) {
	b.updates.Store(updates)
}
func (b *BotServerMock) CleanUpdates() {
	b.updates.Store([]tg_bot.Update{})
}
