-- +goose Up
-- +goose StatementBegin
UPDATE products
SET description = ''
WHERE description IS NULL;

ALTER TABLE products
    ALTER COLUMN description SET DEFAULT '',
    ALTER COLUMN description SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
    ALTER COLUMN description DROP NOT NULL,
    ALTER COLUMN description DROP DEFAULT;
-- +goose StatementEnd
