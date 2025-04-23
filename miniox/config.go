package miniox

import (
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint          string       `schema:"Endpoint minio кластера без указания схемы" validate:"required"`
	Credentials       *Credentials `schema:"Данные для авторизации,secret"`
	Secure            bool         `schema:"Использовать https или нет"`
	UploadFileThreads uint         `schema:"Максимальное количество потоков для загрузки файла"`
}

func (cfg Config) getCredentials() *credentials.Credentials {
	if cfg.Credentials == nil {
		return nil
	}
	return credentials.NewStaticV4(cfg.Credentials.Id, cfg.Credentials.Secret, cfg.Credentials.Token)
}

type Credentials struct {
	Id     string `schema:"Логин,secret"`
	Secret string `schema:"Пароль,secret"`
	Token  string `schema:"Токен,secret"`
}
