package storages

import (
	"context"
	"gw-exchanger/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RepositoryInterface interface {
	GetRates(context.Context) (map[string]float32, error)
	GetRatesForCurrency(ctx context.Context, from, to string) (float32, error)
	Close()
}

type DBPool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

type Repository struct {
	db  DBPool
	lg  logger.Logger
	ctx context.Context
}

const (
	maxconns = 2000
)
