-- +goose Up
CREATE TABLE purchase_requests (
    user_id BIGINT NOT NULL REFERENCES users(id),
    idempotency_key VARCHAR(128) NOT NULL CHECK (char_length(idempotency_key) > 0),
    product_id BIGINT NOT NULL REFERENCES products(id),
    purchase_id BIGINT REFERENCES purchases(id),
    PRIMARY KEY (user_id, idempotency_key)
);

CREATE INDEX idx_purchase_requests_product_id ON purchase_requests(product_id);
CREATE INDEX idx_purchase_requests_purchase_id ON purchase_requests(purchase_id);

-- Проверяем связь в конце транзакции: при резервировании purchase_id еще неизвестен.
-- +goose StatementBegin
CREATE FUNCTION check_purchase_request_completed() RETURNS TRIGGER AS $$
BEGIN
    -- Нужна текущая строка, а не NEW от INSERT, где purchase_id был NULL.
    IF EXISTS (
        SELECT 1 FROM purchase_requests
        WHERE user_id = NEW.user_id
          AND idempotency_key = NEW.idempotency_key
          AND purchase_id IS NULL
    ) THEN
        RAISE EXCEPTION 'purchase request must reference a purchase before commit'
            USING ERRCODE = '23514';
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER purchase_request_completed
AFTER INSERT OR UPDATE ON purchase_requests
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION check_purchase_request_completed();

-- +goose Down
DROP TABLE purchase_requests;
DROP FUNCTION check_purchase_request_completed();
