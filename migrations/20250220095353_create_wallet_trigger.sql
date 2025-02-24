-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_wallet()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallets (user_id, USD, RUB, EUR)
    VALUES (NEW.id, 0.00, 0.00, 0.00);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_user_insert
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_wallet();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS after_user_insert ON users;
DROP FUNCTION IF EXISTS create_wallet();

-- +goose StatementEnd
