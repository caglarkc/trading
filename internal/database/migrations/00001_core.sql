-- +goose Up
CREATE TABLE assets (
 id uuid PRIMARY KEY, symbol text UNIQUE NOT NULL, name text NOT NULL,
 category text NOT NULL, exchange text NOT NULL, aliases jsonb NOT NULL,
 fixed boolean NOT NULL DEFAULT true, execution_blocked boolean NOT NULL DEFAULT false
);
CREATE TABLE wallets (
 id uuid PRIMARY KEY, code text UNIQUE NOT NULL, category text NOT NULL,
 persona_version text NOT NULL DEFAULT 'persona-v1', account_version bigint NOT NULL DEFAULT 0,
 state jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE journal_transactions (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id),
 event_type text NOT NULL, external_event_id text NOT NULL,
 occurred_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(wallet_id,id), UNIQUE(wallet_id,event_type,external_event_id)
);
CREATE TABLE journal_lines (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, wallet_id uuid NOT NULL,
 transaction_id uuid NOT NULL, account_code text NOT NULL,
 amount numeric(28,8) NOT NULL CHECK(amount <> 0), currency text NOT NULL CHECK(currency='USD'),
 FOREIGN KEY(wallet_id,transaction_id) REFERENCES journal_transactions(wallet_id,id)
);
CREATE INDEX journal_lines_wallet ON journal_lines(wallet_id,id);
-- +goose StatementBegin
CREATE FUNCTION reject_journal_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'journal_is_immutable'; END $$;
CREATE TRIGGER immutable_journal BEFORE UPDATE OR DELETE ON journal_transactions FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
CREATE TRIGGER immutable_lines BEFORE UPDATE OR DELETE ON journal_lines FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
CREATE FUNCTION balanced_journal() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE v_sum numeric; v_count bigint;
BEGIN
 SELECT sum(amount),count(*) INTO v_sum,v_count FROM journal_lines WHERE transaction_id=NEW.id AND wallet_id=NEW.wallet_id;
 IF v_count<2 OR v_sum<>0 THEN RAISE EXCEPTION 'unbalanced_journal'; END IF;
 RETURN NEW;
END $$;
CREATE CONSTRAINT TRIGGER journal_balance AFTER INSERT ON journal_transactions DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION balanced_journal();
-- +goose StatementEnd
CREATE TABLE orders (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id), asset_id uuid NOT NULL REFERENCES assets(id),
 state text NOT NULL, body jsonb NOT NULL, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL,
 UNIQUE(wallet_id,id)
);
CREATE INDEX orders_open ON orders(wallet_id,asset_id,id) WHERE state IN ('accepted','waiting','partially_filled');
CREATE TABLE fills (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL, order_id uuid NOT NULL,
 body jsonb NOT NULL, occurred_at timestamptz NOT NULL, UNIQUE(wallet_id,id),
 FOREIGN KEY(wallet_id,order_id) REFERENCES orders(wallet_id,id)
);
CREATE TABLE decision_runs (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id), job_id uuid,
 state text NOT NULL, session_id text, input jsonb, output jsonb, input_hash text, output_hash text,
 model text NOT NULL, reasoning text NOT NULL, started_at timestamptz NOT NULL DEFAULT now(), finished_at timestamptz,
 UNIQUE(wallet_id,id)
);
CREATE TABLE decision_actions (
 wallet_id uuid NOT NULL, decision_id uuid NOT NULL, client_action_id text NOT NULL,
 body_hash text NOT NULL, receipt jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(wallet_id,decision_id,client_action_id), FOREIGN KEY(wallet_id,decision_id) REFERENCES decision_runs(wallet_id,id)
);
CREATE TABLE receipts (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id), decision_id uuid,
 body jsonb NOT NULL, delivered_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(wallet_id,decision_id) REFERENCES decision_runs(wallet_id,id)
);
CREATE INDEX receipts_wallet ON receipts(wallet_id,id);
CREATE TABLE wallet_sessions (
 wallet_id uuid PRIMARY KEY REFERENCES wallets(id), session_id text, previous_session_id text,
 checkpoint_at timestamptz, reset_reason text, updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE memory_versions (
 wallet_id uuid NOT NULL REFERENCES wallets(id), version bigint NOT NULL, decision_id uuid NOT NULL,
 previous_version bigint, summary text NOT NULL CHECK(length(summary)<=12000), created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(wallet_id,version), FOREIGN KEY(wallet_id,decision_id) REFERENCES decision_runs(wallet_id,id),
 FOREIGN KEY(wallet_id,previous_version) REFERENCES memory_versions(wallet_id,version)
);
CREATE TABLE observations (
 id text PRIMARY KEY, asset_id uuid NOT NULL REFERENCES assets(id), source_id text NOT NULL,
 event_time timestamptz, observed_at timestamptz NOT NULL, body jsonb NOT NULL
);
CREATE INDEX observation_time ON observations(asset_id,event_time DESC);
CREATE TABLE market_snapshots (
 id uuid PRIMARY KEY, asset_id uuid NOT NULL REFERENCES assets(id), captured_at timestamptz NOT NULL, body jsonb NOT NULL
);
CREATE INDEX snapshots_time ON market_snapshots(asset_id,captured_at DESC);
CREATE TABLE evidence (
 id uuid PRIMARY KEY, origin_url text NOT NULL, origin_key text NOT NULL, source_url text NOT NULL,
 published_at timestamptz, observed_at timestamptz NOT NULL, status text NOT NULL, body jsonb NOT NULL,
 supersedes uuid REFERENCES evidence(id)
);
CREATE TABLE system_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, kind text NOT NULL, body jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementBegin
DO $$ DECLARE t text; BEGIN
 FOREACH t IN ARRAY ARRAY['wallets','journal_transactions','journal_lines','orders','fills','decision_runs','decision_actions','receipts','wallet_sessions','memory_versions'] LOOP
 EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY',t);
 EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY',t);
 EXECUTE format('CREATE POLICY trusted_access ON %I TO trading_owner,trading_worker USING (true) WITH CHECK (true)',t);
 EXECUTE format('CREATE POLICY agent_scope ON %I TO trading_agent USING (%I = nullif(current_setting(''app.wallet_id'',true),'''')::uuid) WITH CHECK (%I = nullif(current_setting(''app.wallet_id'',true),'''')::uuid)',t, CASE WHEN t='wallets' THEN 'id' ELSE 'wallet_id' END,CASE WHEN t='wallets' THEN 'id' ELSE 'wallet_id' END);
 END LOOP;
 FOREACH t IN ARRAY ARRAY['wallets','journal_transactions','journal_lines','orders','fills','decision_runs','decision_actions','receipts'] LOOP
 EXECUTE format('CREATE POLICY operator_read ON %I FOR SELECT TO trading_operator USING (true)',t);
 END LOOP;
END $$;
-- +goose StatementEnd
GRANT SELECT,INSERT,UPDATE ON ALL TABLES IN SCHEMA public TO trading_worker;
GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA public TO trading_worker;
GRANT SELECT ON assets,observations,market_snapshots,evidence TO trading_agent;
GRANT SELECT,UPDATE ON wallets TO trading_agent;
GRANT SELECT,INSERT ON journal_transactions,journal_lines,fills,decision_actions,receipts,memory_versions TO trading_agent;
GRANT SELECT,INSERT,UPDATE ON orders TO trading_agent;
GRANT SELECT ON decision_runs,wallet_sessions TO trading_agent;
GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA public TO trading_agent;
GRANT SELECT ON assets,wallets,journal_transactions,journal_lines,orders,fills,decision_runs,decision_actions,receipts,observations,market_snapshots,evidence,system_events TO trading_operator;
-- +goose Down
-- Monetary history is never dropped by automatic rollback.
SELECT 1;
