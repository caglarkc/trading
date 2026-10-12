-- +goose Up
CREATE TABLE wallet_watchlist_members (
 wallet_id uuid NOT NULL REFERENCES wallets(id),
 asset_id uuid NOT NULL REFERENCES assets(id),
 review_job_id uuid NOT NULL REFERENCES jobs(id),
 active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(wallet_id,asset_id)
);
ALTER TABLE wallet_watchlist_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE wallet_watchlist_members FORCE ROW LEVEL SECURITY;
CREATE POLICY trusted_watchlist ON wallet_watchlist_members TO trading_owner,trading_worker USING(true) WITH CHECK(true);
CREATE POLICY agent_watchlist ON wallet_watchlist_members FOR SELECT TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid);
CREATE POLICY operator_watchlist ON wallet_watchlist_members FOR SELECT TO trading_operator USING(true);
GRANT SELECT,INSERT,UPDATE ON wallet_watchlist_members TO trading_worker;
GRANT SELECT ON wallet_watchlist_members TO trading_agent,trading_operator;
-- +goose Down
SELECT 1;
