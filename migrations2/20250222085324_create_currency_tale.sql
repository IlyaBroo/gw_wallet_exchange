-- +goose Up
-- +goose StatementBegin

CREATE TABLE currency_rates_usd (
    id SERIAL PRIMARY KEY,
    currency_code VARCHAR(3) NOT NULL,
    exchange_rate REAL NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO currency_rates_usd (currency_code, exchange_rate) VALUES
('USD', 1.00),
('RUB', 0.89),
('EUR', 1.15);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE currency_rates_usd;

-- +goose StatementEnd