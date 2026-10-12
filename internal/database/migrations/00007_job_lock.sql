-- +goose Up
GRANT UPDATE(lease_until) ON jobs TO trading_agent;
CREATE POLICY agent_job_lock ON jobs FOR UPDATE TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid) WITH CHECK(false);
-- +goose Down
SELECT 1;
