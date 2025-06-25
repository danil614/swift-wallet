package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func New() (*Config, error) {
	cfg := &Config{
		HTTPPort: os.Getenv("APP_PORT"),

		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBName: os.Getenv("DB_NAME"),
	}

	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}

	switch {
	case cfg.DBHost == "",
		cfg.DBPort == "",
		cfg.DBUser == "",
		cfg.DBPass == "",
		cfg.DBName == "":
		return nil, errors.New("missing required DB_* environment variables")
	}

	return cfg, nil
}
