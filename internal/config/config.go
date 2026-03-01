package config

import (
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
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:       must("DATABASE_URL"),
		JWTAccessSecret:   must("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:  must("JWT_REFRESH_SECRET"),
		AccessTTLSeconds:  mustInt("ACCESS_TTL_SECONDS", 900),
		RefreshTTLSeconds: mustInt("REFRESH_TTL_SECONDS", 2592000),
		Port:              get("PORT", "8080"),
		MaxDBConnections:  mustInt("MAX_DB_CONNECTIONS", 10),
	}, nil
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func mustInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic("invalid int env " + k)
	}
	return n
}
