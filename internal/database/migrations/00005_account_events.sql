-- +goose Up
CREATE TABLE account_events (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id), event_type text NOT NULL,
 external_event_id text NOT NULL, occurred_at timestamptz NOT NULL,
 state_hash text NOT NULL, UNIQUE(wallet_id,event_type,external_event_id)
);
CREATE TABLE order_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, wallet_id uuid NOT NULL, order_id uuid NOT NULL,
 body jsonb NOT NULL, occurred_at timestamptz NOT NULL,
 FOREIGN KEY(wallet_id,order_id) REFERENCES orders(wallet_id,id)
);
CREATE INDEX order_events_cursor ON order_events(wallet_id,id);
CREATE TABLE settlements (
 wallet_id uuid NOT NULL, fill_id uuid NOT NULL, amount numeric(28,8) NOT NULL CHECK(amount>=0),
 due_at timestamptz NOT NULL, settled boolean NOT NULL DEFAULT false,
 PRIMARY KEY(wallet_id,fill_id), FOREIGN KEY(wallet_id,fill_id) REFERENCES fills(wallet_id,id)
);
-- +goose StatementBegin
DO $$ DECLARE t text; BEGIN
 FOREACH t IN ARRAY ARRAY['account_events','order_events','settlements'] LOOP
 EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY',t);
 EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY',t);
 EXECUTE format('CREATE POLICY trusted_access ON %I TO trading_owner,trading_worker USING(true) WITH CHECK(true)',t);
 EXECUTE format('CREATE POLICY agent_scope ON %I TO trading_agent USING(wallet_id=nullif(current_setting(''app.wallet_id'',true),'''')::uuid) WITH CHECK(wallet_id=nullif(current_setting(''app.wallet_id'',true),'''')::uuid)',t);
 EXECUTE format('CREATE POLICY operator_read ON %I FOR SELECT TO trading_operator USING(true)',t);
 END LOOP;
END $$;
CREATE TRIGGER immutable_account_events BEFORE UPDATE OR DELETE ON account_events FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
CREATE TRIGGER immutable_order_events BEFORE UPDATE OR DELETE ON order_events FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
CREATE TRIGGER immutable_fills BEFORE UPDATE OR DELETE ON fills FOR EACH ROW EXECUTE FUNCTION reject_journal_mutation();
-- +goose StatementEnd
GRANT SELECT,INSERT ON account_events,order_events TO trading_agent,trading_worker;
GRANT SELECT,INSERT,UPDATE ON settlements TO trading_agent,trading_worker;
GRANT SELECT ON account_events,order_events,settlements TO trading_operator;
GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA public TO trading_agent,trading_worker;
-- +goose Down
SELECT 1;
