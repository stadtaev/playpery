package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPAddr string     `env:"HTTP_ADDR" envDefault:":8080"`
	DBPath   string     `env:"DB_PATH,required"`
	LogLevel slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parsing environment: %w", err)
	}
	return &cfg, nil
}
