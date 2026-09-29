package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

func Load(path string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}

	if value := os.Getenv("HTTP_HOST"); value != "" {
		cfg.HTTP.Host = value
	}

	if value := os.Getenv("PORT"); value != "" {
		cfg.HTTP.Port = value
	}

	if value := os.Getenv("DATABASE_HOST"); value != "" {
		cfg.Database.Host = value
	}

	if value := os.Getenv("DATABASE_PORT"); value != "" {
		cfg.Database.Port = value
	}

	if value := os.Getenv("DATABASE_USER"); value != "" {
		cfg.Database.User = value
	}

	if value := os.Getenv("DATABASE_PASSWORD"); value != "" {
		cfg.Database.Password = value
	}

	if value := os.Getenv("DATABASE_NAME"); value != "" {
		cfg.Database.Name = value
	}

	if value := os.Getenv("DATABASE_SSLMODE"); value != "" {
		cfg.Database.SSLMode = value
	}

	if value := os.Getenv("TELEGRAM_TOKEN"); value != "" {
		cfg.Telegram.Token = value
	}

	return &cfg, nil
}
