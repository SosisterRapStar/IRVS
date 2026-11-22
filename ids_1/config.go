package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func LoadConfig() (Config, error) {
	connectionURL, err := databaseURL()
	if err != nil {
		return Config{}, err
	}
	return Config{
		HTTPAddr:    ":" + env("APP_PORT", "8080"),
		DatabaseURL: connectionURL,
	}, nil
}

func databaseURL() (string, error) {
	if connectionURL := os.Getenv("DATABASE_URL"); connectionURL != "" {
		return connectionURL, nil
	}
	password := os.Getenv("DB_PASSWORD")
	if path := os.Getenv("DB_PASSWORD_FILE"); path != "" {
		secret, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read DB_PASSWORD_FILE: %w", err)
		}
		// The PostgreSQL image strips trailing newlines when it reads the secret.
		password = strings.TrimRight(string(secret), "\n")
	}
	name := env("DB_NAME", "app")
	host := strings.Trim(env("DB_HOST", "localhost"), "[]")
	connectionURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), password),
		Host:     net.JoinHostPort(host, env("DB_PORT", "5432")),
		Path:     "/" + name,
		RawPath:  "/" + url.PathEscape(name),
		RawQuery: url.Values{"sslmode": {env("DB_SSLMODE", "disable")}}.Encode(),
	}
	return connectionURL.String(), nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
