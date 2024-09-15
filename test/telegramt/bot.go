package telegramt

import (
	"context"

	"github.com/Falokut/go-kit/telegram_bot"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/httpt"
)

func TestBot(t *test.Test) (*telegram_bot.BotAPI, *BotServerMock) {
	mockServer := httpt.NewMock(t)
	botServerMock := NewBotServerMock(
		t,
		mockServer,
		telegram_bot.User{UserName: "test"},
	)

	bot, err := telegram_bot.NewBotAPIWithAPIEndpoint(
		context.Background(),
		"test",
		formatBotApiEndpoint(mockServer.BaseURL()),
		t.Logger(),
	)
	t.Assert().NoError(err)
	return bot, botServerMock
}
func formatBotApiEndpoint(baseUrl string) string {
	return baseUrl + "/%s/%s"
}
