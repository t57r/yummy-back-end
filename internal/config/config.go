package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	JWTAccessSecret   string
	JWTRefreshSecret  string
	AccessTTLSeconds  int
	RefreshTTLSeconds int
	Port              string
	MaxDBConnections  int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	databaseURL, err := must("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	jwtAccessSecret, err := must("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}
	jwtRefreshSecret, err := must("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}
	accessTTL, err := mustInt("ACCESS_TTL_SECONDS", 900)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := mustInt("REFRESH_TTL_SECONDS", 2592000)
	if err != nil {
		return nil, err
	}
	maxDBConnections, err := mustInt("MAX_DB_CONNECTIONS", 10)
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:       databaseURL,
		JWTAccessSecret:   jwtAccessSecret,
		JWTRefreshSecret:  jwtRefreshSecret,
		AccessTTLSeconds:  accessTTL,
		RefreshTTLSeconds: refreshTTL,
		Port:              get("PORT", "8080"),
		MaxDBConnections:  maxDBConnections,
	}, nil
}

func must(k string) (string, error) {
	v := os.Getenv(k)
	if v == "" {
		return "", fmt.Errorf("missing env: %s", k)
	}
	return v, nil
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func mustInt(k string, def int) (int, error) {
	v := os.Getenv(k)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid int env %s", k)
	}
	return n, nil
}
