-- +goose Up
CREATE TABLE strategy_versions (
 wallet_id uuid NOT NULL REFERENCES wallets(id),
 version bigint NOT NULL CHECK(version>0),
 job_id uuid NOT NULL REFERENCES jobs(id),
 reviewed_through_seq bigint NOT NULL CHECK(reviewed_through_seq>=0),
 input jsonb NOT NULL,
 report jsonb NOT NULL,
 session_id text NOT NULL,
 usage jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(wallet_id,version), UNIQUE(wallet_id,job_id)
);
CREATE TRIGGER immutable_strategy BEFORE UPDATE OR DELETE ON strategy_versions
 FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
ALTER TABLE strategy_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE strategy_versions FORCE ROW LEVEL SECURITY;
CREATE POLICY trusted_strategy ON strategy_versions TO trading_owner,trading_worker USING(true) WITH CHECK(true);
CREATE POLICY agent_strategy ON strategy_versions FOR SELECT TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid);
CREATE POLICY operator_strategy ON strategy_versions FOR SELECT TO trading_operator USING(true);
GRANT SELECT,INSERT ON strategy_versions TO trading_worker;
GRANT SELECT ON strategy_versions TO trading_agent,trading_operator;
-- +goose Down
SELECT 1;
