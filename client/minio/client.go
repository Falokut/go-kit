package minio_client

import (
	"context"
	"io"
	"time"

	"github.com/Falokut/go-kit/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

type Client struct {
	*minio.Client
	uploadThreads uint
}

func NewMinio(cfg config.Minio) (Client, error) {
	cli, err := minio.New(cfg.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Token),
			Secure: cfg.Secure,
		},
	)
	if err != nil {
		return Client{}, errors.WithMessage(err, "new minio client")
	}
	_, err = cli.HealthCheck(time.Second * 5)
	if err != nil {
		return Client{}, errors.WithMessage(err, "init healthcheck minio")
	}
	return Client{
		Client:        cli,
		uploadThreads: cfg.UploadFileThreads,
	}, nil
}

func (m Client) PutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	reader io.Reader,
	objectSize int64,
	opts minio.PutObjectOptions,
) (minio.UploadInfo, error) {
	opts.ConcurrentStreamParts = m.uploadThreads > 1
	opts.NumThreads = m.uploadThreads
	return m.Client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (m Client) HealthCheck(ctx context.Context) error {
	if m.IsOnline() {
		return nil
	}
	return errors.New("minio offline")
}
