package storages

import (
	"context"
	"errors"
	"gw-currency-wallet/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type RepositoryInterface interface {
	Deposit(user_id int, amount decimal.Decimal, currency string, ctx context.Context) error
	Withdraw(user_id int, amount decimal.Decimal, currency string, ctx context.Context) error
	GetBalance(user_id int, ctx context.Context) (Balance, error)
	CheckUser(username string, email string, ctx context.Context) (bool, error)
	AddUser(req RegisterRequest, ctx context.Context) error
	GetUser(username string, ctx context.Context) (User, error)
	ExchangeForCurrency(ctx context.Context, from, to string, amount decimal.Decimal, kurs float32, user_id int) (map[string]decimal.Decimal, error)
	Close()
}

type DBPool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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

var (
	ErrWalletid = errors.New("wallet with this username not found")
	ErrWithdraw = errors.New("insufficient funds or wallet with this username not found")
	ErrExch     = errors.New("func exchangeForCurrency insufficient funds or wallet with this username not found")
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Balance struct {
	USD decimal.Decimal `json:"USD"`
	RUB decimal.Decimal `json:"RUB"`
	EUR decimal.Decimal `json:"EUR"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
