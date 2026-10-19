-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(200) NOT NULL UNIQUE CHECK (char_length(email) > 0),
    name VARCHAR(200) NOT NULL CHECK (char_length(name) > 0),
    balance BIGINT NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    external_id BIGINT NOT NULL UNIQUE CHECK (external_id > 0),
    name VARCHAR(200) NOT NULL CHECK (char_length(name) > 0),
    description TEXT,
    price BIGINT NOT NULL CHECK (price > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE purchases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    price BIGINT NOT NULL CHECK (price > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_purchases ON purchases(user_id DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS purchases;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS products;


-- +goose StatementEnd
