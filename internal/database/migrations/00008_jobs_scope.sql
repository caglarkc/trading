-- +goose Up
-- An explicit forward check covers the staging prototype's already-applied scope.
CREATE UNIQUE INDEX IF NOT EXISTS one_asset_reader ON jobs((payload->>'asset_id')) WHERE kind='price_refresh' AND state='running';
GRANT SELECT ON jobs TO trading_agent;
-- +goose Down
SELECT 1;
