package storages

import (
	"context"
	"fmt"

	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func NewRepository(lg logger.Logger, ctx context.Context, cfg *config.ConfigAdr) RepositoryInterface {
	conf, err := pgxpool.ParseConfig(cfg.Database_url)
	if err != nil {
		lg.FatalCtx(ctx, "Could not parse database URL: ", err)
	}
	conf.MaxConns = maxconns

	pg, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		lg.FatalCtx(ctx, "Could not create connection pool: ", err)
	}
	rep := new(Repository)
	rep.db = pg
	rep.lg = lg
	rep.ctx = ctx
	return rep
}

func (r *Repository) Close() {
	r.db.Close()
}

func (r *Repository) CheckUser(username string, email string, ctx context.Context) (bool, error) {
	var userUsername string
	var userEmail string
	err := r.db.QueryRow(r.ctx, "SELECT username, email FROM users WHERE username = $1 or email = $2 LIMIT 1", username, email).Scan(&userUsername, &userEmail)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.lg.InfoCtx(ctx, "CheckUser no users found")
			return false, nil
		}
		r.lg.ErrorCtx(ctx, "Could not scan user")
		return false, err
	}
	r.lg.InfoCtx(ctx, "username or email exists")
	return true, nil
}

func (r *Repository) AddUser(user RegisterRequest, ctx context.Context) error {

	_, err := r.db.Exec(r.ctx, "INSERT INTO users (username, email, pass) VALUES ($1, $2, $3)", user.Username, user.Email, user.Password)
	if err != nil {
		r.lg.ErrorCtx(ctx, "func adduser sql query failed")
		return err
	}
	r.lg.InfoCtx(ctx, "func adduser sql complete")
	return nil
}

func (r *Repository) GetUser(username string, ctx context.Context) (User, error) {
	user := new(User)
	err := r.db.QueryRow(r.ctx, "SELECT username, pass, id FROM users WHERE username = $1 ", username).Scan(&user.Username, &user.Password, &user.Id)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.lg.InfoCtx(ctx, "GetUser no users found")
			return User{}, err
		}
		r.lg.ErrorCtx(ctx, "Could not scan user")
		return User{}, err
	}
	r.lg.InfoCtx(ctx, "GetUser sql complete")
	return *user, nil
}

func (r *Repository) GetBalance(user_id int, ctx context.Context) (Balance, error) {
	balance := new(Balance)
	r.lg.DebugCtx(ctx, fmt.Sprintf("user_id: %v", user_id))
	err := r.db.QueryRow(r.ctx, "SELECT USD, RUB, EUR FROM wallets WHERE user_id = $1 ", user_id).Scan(&balance.USD, &balance.RUB, &balance.EUR)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.lg.InfoCtx(ctx, "GetBalance no wallet found")
			return Balance{}, err
		}
		r.lg.ErrorCtx(ctx, fmt.Sprintf("Could not scan balance errors: %v", err))
		return Balance{}, err
	}
	r.lg.InfoCtx(ctx, "GetBalance sql complete")
	return *balance, nil
}

func (r *Repository) Deposit(user_id int, amount float64, currency string, ctx context.Context) error {
	queryString := fmt.Sprintf("UPDATE wallets SET %s = %s + $1 WHERE user_id = $2", currency, currency)
	result, err := r.db.Exec(r.ctx, queryString, amount, user_id)
	if err != nil {
		r.lg.ErrorCtx(ctx, "func deposit sql query failed")
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.lg.InfoCtx(ctx, "func deposit wallet with this username not found")
		return ErrWalletid
	}
	r.lg.InfoCtx(ctx, "func deposit sql complete")
	return err
}

func (r *Repository) Withdraw(user_id int, amount float64, currency string, ctx context.Context) error {
	queryString := fmt.Sprintf("UPDATE wallets SET %s = %s - $1 WHERE user_id = $2 AND %s >= $3", currency, currency, currency)
	result, err := r.db.Exec(ctx, queryString, amount, user_id, amount)
	rowsAffected := result.RowsAffected()
	if err != nil {
		r.lg.ErrorCtx(ctx, "func withdraw sql query failed")
		return err
	}
	if rowsAffected == 0 {
		r.lg.InfoCtx(ctx, "func withdraw insufficient funds or wallet with this username not found")
		return ErrWithdraw
	}
	r.lg.InfoCtx(ctx, "func withdraw sql complete")
	return nil
}

func (r *Repository) ExchangeForCurrency(ctx context.Context, from, to string, amount float64, kurs float32, user_id int) (map[string]float64, error) {
	kurs64 := float64(kurs)
	queryString := fmt.Sprintf("UPDATE wallets SET %s = %s - $1, %s = %s + ($2::float * $3::float) WHERE user_id = $4 AND %s >= $5", from, from, to, to, from)
	result, err := r.db.Exec(ctx, queryString, amount, amount, kurs64, user_id, amount)
	if err != nil {
		r.lg.ErrorCtx(ctx, "func exchangeForCurrency sql query failed")
		return nil, err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.lg.InfoCtx(ctx, "func exchangeForCurrency insufficient funds or wallet with this username not found")
		return nil, ErrExch
	}
	res := make(map[string]float64)
	var fromvalue, tovalue float64
	queryString2 := fmt.Sprintf("SELECT %s, %s FROM wallets WHERE user_id = $1", from, to)
	err = r.db.QueryRow(ctx, queryString2, user_id).Scan(&fromvalue, &tovalue)
	if err != nil {
		r.lg.ErrorCtx(ctx, fmt.Sprintf("func exchangeForCurrency scan errors: %v", err))
		return nil, err
	}
	res[from] = fromvalue
	res[to] = tovalue

	r.lg.InfoCtx(ctx, "func exchangeForCurrency sql complete")
	return res, nil
}
