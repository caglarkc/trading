-- +goose Up
CREATE TABLE jobs (
 id uuid PRIMARY KEY, kind text NOT NULL, job_key text NOT NULL, wallet_id uuid REFERENCES wallets(id),
 payload jsonb NOT NULL, depends_on uuid REFERENCES jobs(id), priority integer NOT NULL DEFAULT 10,
 state text NOT NULL DEFAULT 'pending', available_at timestamptz NOT NULL, deadline timestamptz NOT NULL,
 lease_owner text, lease_until timestamptz, fencing_token bigint NOT NULL DEFAULT 0,
 attempts integer NOT NULL DEFAULT 0, last_error text, result jsonb,
 created_at timestamptz NOT NULL DEFAULT now(), finished_at timestamptz,
 UNIQUE(kind,job_key)
);
CREATE INDEX due_jobs ON jobs(priority,available_at,id) WHERE state IN ('pending','running');
CREATE UNIQUE INDEX one_asset_reader ON jobs((payload->>'asset_id')) WHERE kind='price_refresh' AND state='running';
CREATE UNIQUE INDEX one_wallet_writer ON jobs(wallet_id) WHERE wallet_id IS NOT NULL AND state='running';
ALTER TABLE decision_runs ADD CONSTRAINT decision_job_fk FOREIGN KEY(job_id) REFERENCES jobs(id);
ALTER TABLE decision_runs ADD COLUMN fencing_token bigint;
ALTER TABLE decision_runs ADD COLUMN deadline timestamptz;
CREATE TABLE research_bundles (
 id uuid PRIMARY KEY REFERENCES jobs(id), as_of timestamptz NOT NULL, body jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE universe_versions (
 id uuid PRIMARY KEY, week_key text UNIQUE NOT NULL, effective_at timestamptz NOT NULL, body jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE dynamic_members (
 version_id uuid NOT NULL REFERENCES universe_versions(id), asset_id uuid NOT NULL REFERENCES assets(id),
 PRIMARY KEY(version_id,asset_id)
);
CREATE TABLE calendar_events (
 id uuid PRIMARY KEY, category text NOT NULL, event_key text UNIQUE NOT NULL, expected_at timestamptz,
 status text NOT NULL, source_url text NOT NULL, body jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE runtime_control (
 id boolean PRIMARY KEY DEFAULT true CHECK(id), paused boolean NOT NULL DEFAULT false,
 execution_enabled boolean NOT NULL DEFAULT false, live_acceptance jsonb NOT NULL DEFAULT '{}', updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO runtime_control(id) VALUES(true);
GRANT SELECT,INSERT,UPDATE ON jobs,research_bundles,universe_versions,dynamic_members,calendar_events,runtime_control TO trading_worker;
GRANT SELECT ON jobs,research_bundles,universe_versions,dynamic_members,calendar_events,runtime_control TO trading_operator;
GRANT SELECT ON research_bundles,universe_versions,dynamic_members,calendar_events TO trading_agent;
ALTER TABLE jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE jobs FORCE ROW LEVEL SECURITY;
CREATE POLICY trusted_jobs ON jobs TO trading_owner,trading_worker USING(true) WITH CHECK(true);
CREATE POLICY operator_jobs ON jobs FOR SELECT TO trading_operator USING(true);
CREATE POLICY agent_jobs ON jobs FOR SELECT TO trading_agent USING(wallet_id=nullif(current_setting('app.wallet_id',true),'')::uuid);
GRANT SELECT ON jobs TO trading_agent;
-- +goose Down
SELECT 1;
