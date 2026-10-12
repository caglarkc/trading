-- +goose Up
ALTER TABLE assets ADD COLUMN quote_not_before timestamptz;
ALTER TABLE decision_runs ADD COLUMN output_bytes bytea;
-- +goose Down
SELECT 1;
