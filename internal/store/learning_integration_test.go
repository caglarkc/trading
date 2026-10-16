package store_test

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/database"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/marketdata"
	. "github.com/caglarkc/trading/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"os"
	"testing"
	"time"
)

func TestStrategyJournalCoverageAndIsolation(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if err := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); err != nil {
		t.Fatal(err)
	}
	owner, err := Open(ctx, os.Getenv("TEST_OWNER_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	worker, err := Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	agent, err := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Close()
	cal, err := calendar.Load("../../config/calendar.json")
	if err != nil {
		t.Fatal(err)
	}
	fees, err := domain.LoadFees("../../config/fees.json")
	if err != nil {
		t.Fatal(err)
	}
	wallet := WalletID(Codes[0])
	peer := WalletID(Codes[1])
	var asset string
	if err = owner.Pool.QueryRow(ctx, "SELECT id::text FROM assets WHERE symbol='NFLX'").Scan(&asset); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 16, 14, 0, 0, time.Now().Nanosecond(), time.UTC)
	did := uuid.New()
	if _, err = owner.Pool.Exec(ctx, "INSERT INTO decision_runs(id,wallet_id,state,input,model,reasoning,started_at) VALUES($1,$2,'ready','{\"evidence_ids\":[]}','fixture','low',$3)", did, wallet, now); err != nil {
		t.Fatal(err)
	}
	snap := func(price string, at time.Time) marketdata.Snapshot {
		rows := []marketdata.Observation{}
		for _, source := range []string{"stockanalysis", "yahoo", "webull"} {
			rows = append(rows, marketdata.Observation{ID: uuid.NewString(), AssetID: asset, SourceID: source, IdentityVerified: true, Currency: "USD", ObservedAt: at, Price: str(price)})
		}
		return marketdata.Snapshot{Asset: marketdata.Asset{ID: asset}, Observations: rows, ExecutionMode: marketdata.ObservedPriceMode}
	}
	buy := domain.Action{Action: "SUBMIT", ClientID: "learning-buy", AssetID: &asset, Side: str("BUY"), Quantity: qty(2), Type: str("MARKET"), TIF: str("DAY"), Budget: str("200"), Thesis: "Original entry assumption recorded before the outcome."}
	receipt, err := agent.ApplyAction(ctx, wallet.String(), did.String(), buy, now, cal, fees)
	if err != nil || receipt.Order == nil {
		t.Fatal(err, receipt)
	}
	receipt, err = worker.Execute(ctx, wallet.String(), receipt.Order.ID, snap("80", now.Add(time.Second)), now.Add(time.Second), cal, fees, true)
	if err != nil || receipt.Fill == nil {
		t.Fatal(err, receipt)
	}
	sell := domain.Action{Action: "SUBMIT", ClientID: "learning-sell", AssetID: &asset, Side: str("SELL"), Quantity: qty(1), Type: str("MARKET"), TIF: str("DAY"), Thesis: "A partial exit is a closed portion and must be reviewed."}
	receipt, err = agent.ApplyAction(ctx, wallet.String(), did.String(), sell, now.Add(2*time.Second), cal, fees)
	if err != nil || receipt.Order == nil {
		t.Fatal(err, receipt)
	}
	receipt, err = worker.Execute(ctx, wallet.String(), receipt.Order.ID, snap("70", now.Add(3*time.Second)), now.Add(3*time.Second), cal, fees, true)
	if err != nil || receipt.Fill == nil {
		t.Fatal(err, receipt)
	}
	fills, err := worker.LearningFills(ctx, wallet.String(), 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	var through int64
	for _, f := range fills {
		through = f.Seq
		if f.ID != receipt.Fill.ID {
			continue
		}
		found = true
		var facts struct {
			PNL  string `json:"net_realized_pnl_usd"`
			Cost string `json:"disposed_cost_basis_usd"`
		}
		if err = json.Unmarshal(f.Facts, &facts); err != nil {
			t.Fatal(err)
		}
		if !decimal.RequireFromString(facts.PNL).IsNegative() || !decimal.RequireFromString(facts.Cost).IsPositive() {
			t.Fatal("journal outcome missing", string(f.Facts))
		}
	}
	if !found {
		t.Fatal("partial exit not reviewed")
	}
	later, err := worker.LearningFills(ctx, wallet.String(), through)
	if err != nil || len(later) != 0 {
		t.Fatal("watermark replay", err, later)
	}
	q := jobs.Queue{Pool: worker.Pool}
	kind := "fixture-learning-" + uuid.NewString()
	jid, err := q.Enqueue(ctx, jobs.Job{Kind: kind, Key: uuid.NewString(), Wallet: &wallet, Available: time.Now().Add(-time.Second), Deadline: time.Now().Add(time.Minute)}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	j, err := q.Acquire(ctx, "learning-test", []string{kind})
	if err != nil || j == nil || j.ID != jid {
		t.Fatal(err, j)
	}
	empty := StrategyState{Report: json.RawMessage(`{}`)}
	report := json.RawMessage(`{"strategy_summary":"fixture"}`)
	if err = worker.SaveStrategy(ctx, peer.String(), empty, through, report, report, "fixture", nil, *j); err == nil {
		t.Fatal("cross-wallet strategy write allowed")
	}
	if err = worker.SaveStrategy(ctx, wallet.String(), empty, through, report, report, "fixture", nil, *j); err != nil {
		t.Fatal(err)
	}
	if err = worker.SaveStrategy(ctx, wallet.String(), empty, through, report, report, "fixture", nil, *j); err == nil {
		t.Fatal("duplicate revision accepted")
	}
	state, err := worker.LatestStrategy(ctx, wallet.String())
	if err != nil || state.Version != 1 || state.Through != through {
		t.Fatal(state, err)
	}
	err = agent.Scoped(ctx, peer.String(), func(tx pgx.Tx, _ *dbgen.Queries) error {
		var n int
		err := tx.QueryRow(ctx, "SELECT count(*) FROM strategy_versions WHERE wallet_id=$1", wallet).Scan(&n)
		if err == nil && n != 0 {
			t.Fatal("peer strategy leaked")
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = worker.Pool.Exec(ctx, "UPDATE strategy_versions SET report='{}' WHERE wallet_id=$1", wallet); err == nil {
		t.Fatal("immutable strategy changed")
	}
	stale := *j
	stale.Fence++
	if err = worker.SaveStrategy(ctx, wallet.String(), state, through, report, report, "fixture", nil, stale); err == nil {
		t.Fatal("stale job wrote a strategy")
	}
	if err = q.Finish(ctx, *j, map[string]bool{"tested": true}, nil); err != nil {
		t.Fatal(err)
	}
}
