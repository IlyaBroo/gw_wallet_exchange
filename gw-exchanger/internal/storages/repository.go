package storages

import (
	"context"
	"fmt"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *Repository) GetRates(ctx context.Context) (map[string]float32, error) {

	rates := make(map[string]float32)

	rows, err := r.db.Query(ctx, "SELECT currency_code, exchange_rate FROM currency_rates_usd")
	if err != nil {
		r.lg.ErrorCtx(ctx, "func get_rates sql query failed")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var currencyCode string
		var exchangeRate float32

		if err := rows.Scan(&currencyCode, &exchangeRate); err != nil {
			r.lg.ErrorCtx(ctx, fmt.Sprintf("Error scanning row: %v ", err))
			return nil, err
		}

		rates[currencyCode] = exchangeRate
	}
	if err := rows.Err(); err != nil {
		r.lg.ErrorCtx(ctx, fmt.Sprintf("Error iterating rows: %v ", err))
		return nil, err
	}

	r.lg.InfoCtx(ctx, fmt.Sprintf("take rows: %v", rates))
	return rates, nil
}

func (r *Repository) GetRatesForCurrency(ctx context.Context, from, to string) (float32, error) {

	rates := make(map[string]float32)

	rows, err := r.db.Query(ctx, "SELECT currency_code, exchange_rate FROM currency_rates_usd WHERE currency_code = $1 OR currency_code = $2 LIMIT 2", from, to)
	if err != nil {
		r.lg.ErrorCtx(ctx, "func get_rates sql query failed")
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var currencyCode string
		var exchangeRate float32

		if err := rows.Scan(&currencyCode, &exchangeRate); err != nil {
			r.lg.ErrorCtx(ctx, fmt.Sprintf("Error scanning row: %v ", err))
			return 0, err
		}

		rates[currencyCode] = exchangeRate
	}
	if err := rows.Err(); err != nil {
		r.lg.ErrorCtx(ctx, fmt.Sprintf("Error iterating rows: %v ", err))
		return 0, err
	}
	var res float32
	res = 1 / rates[to] * rates[from]
	// res = roundToTwoDecimalPlaces(res)
	r.lg.InfoCtx(ctx, fmt.Sprintf("take rows: %v", rates))
	return res, nil

}

// func roundToTwoDecimalPlaces(value float32) float32 {
// 	return float32(math.Round(float64(value)*100) / 100)
// }
