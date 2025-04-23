package tg_botx

type Config struct {
	Token        string `schema:"Токен для бота,secret" validate:"required,min=1"`
	PaymentToken string `schema:"Токен для интеграции с платёжной платформой,secret"`

	// timeout in seconds
	Timeout        int      `schema:"Время получения сообщений от сервера" validate:"required,min=1,max=1000"`
	Limit          int      `schema:"Количество читаемых сообщений за раз" validate:"required,min=1,max=1000"`
	AllowedUpdates []string `schema:"Типы сообщений для получения ботом"`

	ApiEndpoint string

	RetryDelaySec int `schema:"Время через которое повторится обработка сообщений, в секундах"` // default = 3
	MaxRetryCount int `schema:"Максимальное количество повторений обработки сообщений"`         // -1 - infinity retry count
}
