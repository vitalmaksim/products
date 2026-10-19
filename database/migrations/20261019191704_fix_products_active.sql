-- +goose Up
ALTER TABLE public.products
ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE public.products
DROP COLUMN IF EXISTS active;