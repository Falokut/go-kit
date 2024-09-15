package telegramt

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/telegram_bot"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/httpt"
)

type BotServerMock struct {
	*httpt.MockServer
	updates           atomic.Value
	lastSendedMessage atomic.Value
}

func NewBotServerMock(test *test.Test, mock *httpt.MockServer, botInfo telegram_bot.User) *BotServerMock {
	botServerMock := &BotServerMock{
		MockServer:        mock,
		updates:           atomic.Value{},
		lastSendedMessage: atomic.Value{},
	}
	botServerMock.POST("/test/getMe", func() telegram_bot.APIResponse {
		result, err := json.Marshal(botInfo)
		test.Assert().NoError(err)
		return telegram_bot.APIResponse{
			Ok:     true,
			Result: result,
		}
	})
	botServerMock.POST("/test/getUpdates", func(r *http.Request) telegram_bot.APIResponse {
		if r.URL.Query().Has("timeout") {
			timeoutVal := r.URL.Query().Get("timeout")
			timeout, err := strconv.Atoi(timeoutVal)
			test.Assert().NoError(err)
			time.Sleep(time.Duration(timeout) * time.Second)
		}
		updates := botServerMock.updates.Load().([]telegram_bot.Update)
		offset := 0
		var err error
		if r.URL.Query().Has("offset") {
			offsetVal := r.URL.Query().Get("offset")
			offset, err = strconv.Atoi(offsetVal)
			test.Assert().NoError(err)
		}
		if offset > len(updates) {
			updates = []telegram_bot.Update{}
		} else {
			updates = updates[offset:]
		}

		result, err := json.Marshal(updates)
		test.Assert().NoError(err)
		return telegram_bot.APIResponse{
			Ok:     true,
			Result: result,
		}
	})
	botServerMock.POST("/test/setMyCommands", func() telegram_bot.APIResponse {
		return telegram_bot.APIResponse{Ok: true}
	})
	botServerMock.POST("/test/deleteMyCommands", func() telegram_bot.APIResponse {
		return telegram_bot.APIResponse{Ok: true}
	})
	botServerMock.POST("/test/sendMessage", func(r *http.Request) telegram_bot.APIResponse {
		botServerMock.lastSendedMessage.Store(r.URL.Query())
		return telegram_bot.APIResponse{Ok: true}
	})
	return botServerMock
}

func (b *BotServerMock) SetUpdates(updates []telegram_bot.Update) {
	b.updates.Store(updates)
}
func (b *BotServerMock) CleanUpdates() {
	b.updates.Store([]telegram_bot.Update{})
}
