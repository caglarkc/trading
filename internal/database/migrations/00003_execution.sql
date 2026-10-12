-- +goose Up
CREATE TABLE position_lots (
 id uuid PRIMARY KEY, wallet_id uuid NOT NULL REFERENCES wallets(id), asset_id uuid NOT NULL REFERENCES assets(id),
 fill_id uuid NOT NULL, quantity bigint NOT NULL CHECK(quantity>=0), cost_basis numeric(28,8) NOT NULL CHECK(cost_basis>=0), acquired_at timestamptz NOT NULL,
 FOREIGN KEY(wallet_id,fill_id) REFERENCES fills(wallet_id,id), UNIQUE(wallet_id,id)
);
CREATE INDEX lots_fifo ON position_lots(wallet_id,asset_id,acquired_at,id);
CREATE TABLE reservations (
 wallet_id uuid NOT NULL, order_id uuid NOT NULL, oco_id uuid,
 cash_usd numeric(28,8) NOT NULL CHECK(cash_usd>=0), shares bigint NOT NULL CHECK(shares>=0),
 PRIMARY KEY(wallet_id,order_id), FOREIGN KEY(wallet_id,order_id) REFERENCES orders(wallet_id,id)
);
CREATE TABLE execution_quotes (
 id text PRIMARY KEY, asset_id uuid NOT NULL REFERENCES assets(id), body jsonb NOT NULL, observed_at timestamptz NOT NULL
);
CREATE TABLE quote_consumptions (
 wallet_id uuid NOT NULL REFERENCES wallets(id), quote_id text NOT NULL REFERENCES execution_quotes(id),
 side text NOT NULL, quantity bigint NOT NULL CHECK(quantity>0), PRIMARY KEY(wallet_id,quote_id,side)
);
CREATE TABLE corporate_events (
 id uuid PRIMARY KEY, asset_id uuid NOT NULL REFERENCES assets(id), event_type text NOT NULL,
 effective_at timestamptz NOT NULL, status text NOT NULL, source_url text NOT NULL, body jsonb NOT NULL,
 supersedes uuid REFERENCES corporate_events(id)
);
CREATE TABLE receivables (
 wallet_id uuid NOT NULL REFERENCES wallets(id), event_id uuid NOT NULL REFERENCES corporate_events(id),
 gross numeric(28,8) NOT NULL CHECK(gross>=0), tax numeric(28,8), net numeric(28,8), quantity bigint NOT NULL,
 pay_at timestamptz, status text NOT NULL, PRIMARY KEY(wallet_id,event_id)
);
-- +goose StatementBegin
DO $$ DECLARE t text;BEGIN
 FOREACH t IN ARRAY ARRAY['position_lots','reservations','quote_consumptions','receivables'] LOOP
 EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY',t);
 EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY',t);
 EXECUTE format('CREATE POLICY trusted_access ON %I TO trading_owner,trading_worker USING (true) WITH CHECK (true)',t);
 EXECUTE format('CREATE POLICY agent_scope ON %I TO trading_agent USING (wallet_id=nullif(current_setting(''app.wallet_id'',true),'''')::uuid) WITH CHECK (wallet_id=nullif(current_setting(''app.wallet_id'',true),'''')::uuid)',t);
 EXECUTE format('CREATE POLICY operator_read ON %I FOR SELECT TO trading_operator USING (true)',t);
 END LOOP;
END $$;
-- +goose StatementEnd
GRANT SELECT,INSERT,UPDATE ON position_lots,reservations,quote_consumptions TO trading_agent,trading_worker;
GRANT SELECT,INSERT ON execution_quotes TO trading_agent,trading_worker;
GRANT SELECT ON corporate_events,receivables TO trading_agent;
GRANT SELECT,INSERT,UPDATE ON corporate_events,receivables TO trading_worker;
GRANT SELECT ON position_lots,reservations,quote_consumptions,execution_quotes,corporate_events,receivables TO trading_operator;
-- +goose Down
SELECT 1;
