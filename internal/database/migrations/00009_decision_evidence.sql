-- +goose Up
CREATE TABLE decision_evidence (
 wallet_id uuid NOT NULL, decision_id uuid NOT NULL, evidence_id uuid NOT NULL REFERENCES evidence(id),
 seen_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(wallet_id,decision_id,evidence_id),
 FOREIGN KEY(wallet_id,decision_id) REFERENCES decision_runs(wallet_id,id)
);
ALTER TABLE decision_evidence ENABLE ROW LEVEL SECURITY;
ALTER TABLE decision_evidence FORCE ROW LEVEL SECURITY;
CREATE POLICY trusted_access ON decision_evidence TO trading_owner,trading_worker USING(true) WITH CHECK(true);
CREATE POLICY agent_scope ON decision_evidence FOR SELECT TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid);
CREATE POLICY operator_read ON decision_evidence FOR SELECT TO trading_operator USING(true);
GRANT SELECT ON decision_evidence TO trading_agent,trading_operator;
GRANT SELECT,INSERT ON decision_evidence TO trading_worker;
-- +goose Down
SELECT 1;
