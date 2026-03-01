package db

import (
	"context"

	"yummy/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(cfg *config.Config) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	dbConfig.MaxConns = int32(cfg.MaxDBConnections)
	return pgxpool.NewWithConfig(context.Background(), dbConfig)
}
