-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    pass VARCHAR(255) NOT NULL
);

CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    USD FLOAT NOT NULL,
    RUB FLOAT NOT NULL,
    EUR FLOAT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE wallets;
DROP TABLE users;
-- +goose StatementEnd
