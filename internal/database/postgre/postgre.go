package postgre

import (
	"blog/pkg/consts/errors"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgreConfig struct {
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"123"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
}

type DB struct {
	*sql.DB
}

func NewDB(DBName string, config PostgreConfig, ctx context.Context) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.ErrFailedOpenDB
	}

	var exists bool
	err = db.QueryRow(`SELECT EXISTS (SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`, DBName).Scan(&exists)
	if err != nil {
		return nil, errors.ErrFailedCheckDBExists
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, DBName))
		if err != nil {
			return nil, errors.ErrFailedCreateDB
		}
	}

	db, err = sql.Open("postgres", fmt.Sprintf("%s dbname=%s", dsn, DBName))
	if err != nil {
		return nil, errors.ErrFailedOpenDB
	}

	_, err = db.Conn(ctx)
	if err != nil {
		return nil, errors.ErrFailedConnectDB
	}
	return &DB{db}, nil
}
