-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN active BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN active;
-- +goose StatementEnd
