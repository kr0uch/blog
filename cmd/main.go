package main

import (
	_ "blog/docs"
	"blog/internal/config"
	"blog/internal/database/migrations"
	"blog/internal/database/postgre"
	"blog/internal/logger"
	"blog/internal/storage/minio"
	"blog/internal/transport/rest/servers"
	"context"
	"log"
)

// @title Blog API
// @version 1.0
// @description API для блога с аутентификацией
// @host localhost:8080
// @BasePath /
func main() {
	ctx := context.Background()
	zapLogger := logger.NewLogger()
	ctx = context.WithValue(ctx, logger.LoggerKey, zapLogger)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgre.NewDB(cfg.PostgreConfig.DBName, cfg.PostgreConfig, ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = migrations.Up(db.DB)
	if err != nil {
		log.Fatal(err)
	}

	minioClient, err := minio.NewMinioClient(cfg.MinioClientConfig)
	if err != nil {
		log.Fatal(err)
	}

	server, err := servers.NewBlogServer(cfg.BlogServerConfig, minioClient, db, zapLogger)
	if err != nil {
		log.Fatal(err)
	}
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
