package store

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/database"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"os"
	"testing"
)

func TestPostgresIsolationAndCapital(t *testing.T) {
	dsn := os.Getenv("TEST_OWNER_URL")
	if dsn == "" {
		t.Skip("real isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, dsn); e != nil {
		t.Fatal(e)
	}
	owner, e := Open(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Close()
	data, e := os.ReadFile("../../config/assets.json")
	if e != nil {
		t.Fatal(e)
	}
	assets, e := marketdata.DecodeAssets(data)
	if e != nil {
		t.Fatal(e)
	}
	for range 2 {
		if e = owner.Seed(ctx, assets); e != nil {
			t.Fatal(e)
		}
	}
	var count int
	var capital string
	if e = owner.Pool.QueryRow(ctx, "SELECT count(*),sum((state->>'available_cash_usd')::numeric)::text FROM wallets").Scan(&count, &capital); e != nil || count != 8 || capital != "800000" {
		t.Fatalf("capital: %d %s %v", count, capital, e)
	}
	agent, e := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer agent.Close()
	a, b := WalletID(Codes[0]), WalletID(Codes[1])
	if rows, e := agent.Queries.ListWallets(ctx); e != nil || len(rows) != 0 {
		t.Fatal("unscoped RLS", len(rows), e)
	}
	e = agent.Scoped(ctx, a.String(), func(tx pgx.Tx, q *dbgen.Queries) error {
		rows, e := q.ListWallets(ctx)
		if e != nil || len(rows) != 1 || rows[0].ID != a {
			t.Fatal("scope leak", rows, e)
		}
		if _, e = q.GetWallet(ctx, b); e != pgx.ErrNoRows {
			t.Fatal("peer read", e)
		}
		_, e = tx.Exec(ctx, "INSERT INTO orders(id,wallet_id,asset_id,state,body,created_at,updated_at) VALUES($1,$2,$3,'accepted','{}',now(),now())", uuid.New(), b, assets[0].ID)
		if e == nil {
			t.Fatal("peer write permitted")
		}
		return e
	})
	if e == nil {
		t.Fatal("expected rollback")
	}
	// Cross-wallet composite parent is rejected even for a trusted role.
	tx, e := owner.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx.Rollback(ctx)
	order := uuid.New()
	_, e = tx.Exec(ctx, "INSERT INTO orders(id,wallet_id,asset_id,state,body,created_at,updated_at) VALUES($1,$2,$3,'accepted','{}',now(),now())", order, a, assets[0].ID)
	if e != nil {
		t.Fatal(e)
	}
	_, e = tx.Exec(ctx, "INSERT INTO fills(id,wallet_id,order_id,body,occurred_at) VALUES($1,$2,$3,'{}',now())", uuid.New(), b, order)
	if e == nil {
		t.Fatal("cross wallet FK permitted")
	}
	tx.Rollback(ctx)
	// A one-sided journal must fail at commit, and changing history is forbidden.
	tx, e = owner.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer tx.Rollback(ctx)
	jid := uuid.New()
	_, e = tx.Exec(ctx, "INSERT INTO journal_transactions(id,wallet_id,event_type,external_event_id,occurred_at) VALUES($1,$2,'test',$1::uuid::text,now())", jid, a)
	if e != nil {
		t.Fatal(e)
	}
	_, e = tx.Exec(ctx, "INSERT INTO journal_lines(wallet_id,transaction_id,account_code,amount,currency) VALUES($1,$2,'cash_available',1,'USD')", a, jid)
	if e != nil {
		t.Fatal(e)
	}
	if e = tx.Commit(ctx); e == nil {
		t.Fatal("unbalanced journal committed")
	}
	if _, e = owner.Pool.Exec(ctx, "UPDATE journal_lines SET amount=amount+1 WHERE wallet_id=$1", a); e == nil {
		t.Fatal("journal mutated")
	}
	operator, e := Open(ctx, os.Getenv("TEST_OPERATOR_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer operator.Close()
	if _, e = operator.Pool.Exec(ctx, "UPDATE wallets SET state=state WHERE id=$1", a); e == nil {
		t.Fatal("operator wrote account")
	}
	if _, e = operator.Pool.Exec(ctx, "SELECT * FROM memory_versions"); e == nil {
		t.Fatal("operator read private memory")
	}
	out, _ := json.Marshal(map[string]any{"wallets": count, "capital_usd": capital, "unscoped_rls": "denied", "peer_read_write": "denied", "cross_wallet_fk": "denied", "unbalanced_journal": "denied", "journal_update": "denied", "operator_write": "denied"})
	t.Log(string(out))
}
