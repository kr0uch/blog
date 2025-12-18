package minio

import (
	"blog/pkg/consts/errors"
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClientConfig struct {
	Endpoint string `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	User     string `env:"MINIO_ROOT_USER" env-default:"minioadmin"`
	Password string `env:"MINIO_ROOT_PASSWORD" env-default:"miniopassword"`
	UseSSL   bool   `env:"MINIO_USE_SSL" env-default:"false"`
	Bucket   string `env:"MINIO_BUCKET" env-default:"data"`
}

type MinioClient struct {
	Client *minio.Client
	Cfg    MinioClientConfig
	Bucket string
}

func NewMinioClient(cfg MinioClientConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.User, cfg.Password, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{
		Client: client,
		Cfg:    cfg,
		Bucket: cfg.Bucket,
	}, nil
}

func (r *MinioClient) Upload(ctx context.Context, bucket, filename string, file io.Reader, size int64) (string, error) {
	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return "", errors.ErrMinioBucketNotExists
	}
	if !exists {
		if err = r.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return "", errors.ErrMinioMakeBucket
		}
	}
	info, err := r.Client.PutObject(ctx, bucket, filename, file, size, minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return "", errors.ErrMinioPutObject
	}
	return info.Key, nil
}

func (r *MinioClient) GenerateURL(ctx context.Context, bucket, filename string, expires time.Duration) (string, error) {
	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return "", errors.ErrMinioBucketNotExists
	}
	if !exists {
		return "", errors.ErrMinioBucketNotExists
	}

	url, err := r.Client.PresignedGetObject(ctx, bucket, filename, expires, nil)
	if err != nil {
		return "", errors.ErrMinioPresignedGetObject
	}

	return url.String(), nil
}

func (r *MinioClient) DeleteImage(ctx context.Context, bucket, filename string) error {
	exists, err := r.Client.BucketExists(ctx, bucket)
	if err != nil {
		return errors.ErrMinioBucketNotExists
	}
	if !exists {
		return errors.ErrMinioBucketNotExists
	}

	image, err := r.Client.GetObject(ctx, bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return errors.ErrMinioGetObject
	}
	if image == nil {
		return errors.ErrInvalidImageId
	}

	err = r.Client.RemoveObject(ctx, bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.ErrMinioRemoveObject
	}
	return nil
}
