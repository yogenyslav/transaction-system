package database

import (
	"accountservice/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var maxConnectionAttempts = 5

func MustNewPostgres(cfg *config.Config, retryTimeout int) *pgxpool.Pool {
	var (
		err        error
		pgCfg      *pgxpool.Config
		pool       *pgxpool.Pool
		message    string
		attempts   = 0
		url        = "postgres://%s:%s@%s:5432/%s?sslmode=disable"
		connString = fmt.Sprintf(url, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Db)
	)

	ctx := context.Background()
	for {
		time.Sleep(time.Second * time.Duration(retryTimeout))
		attempts++

		pgCfg, err = pgxpool.ParseConfig(connString)
		if err != nil {
			message = "failed parsing pg config"
			slog.Error(message)
			if attempts > maxConnectionAttempts {
				break
			}
			continue
		}

		pool, err = pgxpool.NewWithConfig(ctx, pgCfg)
		if err != nil {
			message = "failed connecting to pg"
			slog.Error(message)
			if attempts > maxConnectionAttempts {
				break
			}
			continue
		}

		err = pool.Ping(ctx)
		if err != nil {
			message = "failed accessing pg"
			slog.Error(message)
			if attempts > maxConnectionAttempts {
				break
			}
			continue
		}

		return pool
	}

	panic(err)
}
