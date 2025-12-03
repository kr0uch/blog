package config

import (
	"blog/internal/database/postgre"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	postgre.PostgreConfig
}

func NewConfig() (*Config, error) {
	cfg := Config{}
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
