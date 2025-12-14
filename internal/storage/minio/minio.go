package minio

import (
	"blog/pkg/consts/errors"
	"context"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClientConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT" env-default:"http://localhost:9000"`
	AccessKey string `env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`
	SecretKey string `env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
	UseSSL    bool   `env:"MINIO_USE_SSL" env-default:"false"`
}

type MinioClient struct {
	Client *minio.Client
	Cfg    MinioClientConfig
}

func NewMinioClient(cfg MinioClientConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &MinioClient{
		Client: client,
		Cfg:    cfg,
	}, nil
}

func (r *MinioClient) Upload(ctx context.Context, bucket, filename string, file io.Reader, size int64) (string, error) {
	//TODO: проверка что имага существует

	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		if err = r.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return "", err
		}
	}
	info, err := r.Client.PutObject(ctx, bucket, filename, file, size, minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return "", err
	}
	return info.Key, nil
}

func (r *MinioClient) GenerateURL(ctx context.Context, bucket, filename string, expires time.Duration) (string, error) {
	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.ErrMinioBucketNotExists
	}

	url, err := r.Client.PresignedGetObject(ctx, bucket, filename, expires, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (r *MinioClient) DeleteImage(ctx context.Context, bucket, filename string) error {
	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return errors.ErrMinioBucketNotExists
	}

	image, err := r.Client.GetObject(ctx, bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	if image == nil {
		return errors.ErrInvalidImageId
	}

	err = r.Client.RemoveObject(ctx, bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
