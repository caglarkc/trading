-- +goose Up
CREATE TABLE oco_groups (
 wallet_id uuid NOT NULL REFERENCES wallets(id), id uuid NOT NULL, asset_id uuid NOT NULL REFERENCES assets(id),
 reserved_shares bigint NOT NULL CHECK(reserved_shares>=0), PRIMARY KEY(wallet_id,id)
);
ALTER TABLE oco_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE oco_groups FORCE ROW LEVEL SECURITY;
CREATE POLICY trusted_access ON oco_groups TO trading_owner,trading_worker USING(true) WITH CHECK(true);
CREATE POLICY agent_scope ON oco_groups TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid) WITH CHECK(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid);
CREATE POLICY operator_read ON oco_groups FOR SELECT TO trading_operator USING(true);
GRANT SELECT,INSERT,UPDATE ON oco_groups TO trading_agent,trading_worker;
GRANT SELECT ON oco_groups TO trading_operator;
ALTER TABLE reservations ADD CONSTRAINT reservation_oco_wallet FOREIGN KEY(wallet_id,oco_id) REFERENCES oco_groups(wallet_id,id);
ALTER TABLE wallets ADD CONSTRAINT nonnegative_cash CHECK(
 (state->>'available_cash_usd')::numeric>=0 AND (state->>'reserved_cash_usd')::numeric>=0 AND (state->>'unsettled_cash_usd')::numeric>=0
);
-- +goose Down
SELECT 1;
