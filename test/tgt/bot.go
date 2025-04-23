package tgt

import (
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/httpt"
	"github.com/Falokut/go-kit/tg_bot"
	"github.com/Falokut/go-kit/tg_botx"
)

func TestBot(t *test.Test) (*tg_botx.Bot, *BotServerMock) {
	mockServer := httpt.NewMock(t)
	botServerMock := NewBotServerMock(
		t,
		mockServer,
		tg_bot.User{UserName: "test"},
	)

	bot := tg_botx.New(t.Logger())
	err := bot.UpgradeConfig(
		t.T().Context(),
		tg_botx.Config{
			Token:       "test",
			ApiEndpoint: formatBotApiEndpoint(mockServer.BaseURL()),
		},
	)
	t.Assert().NoError(err)
	go bot.Serve(t.T().Context()) // nolint:errcheck

	t.T().Cleanup(func() { bot.Shutdown() })

	return bot, botServerMock
}
func formatBotApiEndpoint(baseUrl string) string {
	return baseUrl + "/%s/%s"
}
