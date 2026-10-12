-- +goose Up
-- Supports an already deployed first staging migration as well as fresh installs.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS seq bigint GENERATED ALWAYS AS IDENTITY;
ALTER TABLE fills ADD COLUMN IF NOT EXISTS seq bigint GENERATED ALWAYS AS IDENTITY;
CREATE UNIQUE INDEX IF NOT EXISTS orders_seq ON orders(seq);
CREATE UNIQUE INDEX IF NOT EXISTS fills_seq ON fills(seq);
GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA public TO trading_worker,trading_agent;
-- +goose Down
SELECT 1;
