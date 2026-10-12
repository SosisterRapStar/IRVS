package main

import (
	"fmt"
	"os"
)

// Config собирается только из переменных окружения — так приложение
// одинаково запускается и локально, и в контейнере.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:    ":" + env("APP_PORT", "8080"),
		DatabaseURL: databaseURL(),
	}
}

// databaseURL: либо готовый DATABASE_URL, либо собираем его из частей.
func databaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env("DB_USER", "app"),
		env("DB_PASSWORD", "app"),
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_NAME", "app"),
		env("DB_SSLMODE", "disable"),
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
