-- +goose Up
-- Таблица и индекс по user_id уже созданы в 00001_init_schema.sql.
-- Индекс по второму внешнему ключу нужен для проверки ссылок при изменении товара.
CREATE INDEX idx_purchases_product_id ON purchases(product_id);

-- +goose Down
DROP INDEX idx_purchases_product_id;
