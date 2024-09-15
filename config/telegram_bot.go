package config

type TelegramBotConfig struct {
	Token        string `yaml:"token" env:"TG_BOT_TOKEN"`
	PaymentToken string `yaml:"payment_token" env:"TG_BOT_PAYMENT_TOKEN"`
	// timeout in seconds
	Timeout int `yaml:"timeout" env:"TG_BOT_TIMEOUT"`
	Limit   int `yaml:"limit" env:"TG_BOT_LIMIT"`
}
