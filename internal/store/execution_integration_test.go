package store_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/database"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/marketdata"
	. "github.com/caglarkc/trading/internal/store"
	"github.com/google/uuid"
	"os"
	"sync"
	"testing"
	"time"
)

func str(s string) *string { return &s }
func qty(n int64) *int64   { return &n }
func TestPostgresOrdersAndExecution(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); e != nil {
		t.Fatal(e)
	}
	owner, e := Open(ctx, os.Getenv("TEST_OWNER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Close()
	worker, e := Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	agent, e := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer agent.Close()
	cal, e := calendar.Load("../../config/calendar.json")
	if e != nil {
		t.Fatal(e)
	}
	var fee domain.FeeProfile
	fb, e := os.ReadFile("../../config/fees.json")
	if e != nil {
		t.Fatal(e)
	}
	json.Unmarshal(fb, &fee)
	now := time.Date(2026, 10, 12, 14, 0, 0, time.Now().Nanosecond(), time.UTC)
	wallet := WalletID(Codes[0]).String()
	asset := uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:asset:AAPL")).String()
	decision := uuid.NewString()
	_, e = owner.Pool.Exec(ctx, "INSERT INTO decision_runs(id,wallet_id,state,input,model,reasoning,started_at) VALUES($1,$2,'ready','{\"evidence_ids\":[]}','fixture','low',$3)", decision, wallet, now)
	if e != nil {
		t.Fatal(e)
	}
	action := func(id, budget string, q int64) domain.Action {
		return domain.Action{Action: "SUBMIT", ClientID: id, AssetID: &asset, Side: str("BUY"), Quantity: qty(q), Type: str("MARKET"), TIF: str("DAY"), Budget: &budget}
	}
	// Serialized wallet row protects against simultaneous settled cash use.
	var wg sync.WaitGroup
	results := make([]domain.Receipt, 2)
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = agent.ApplyAction(ctx, wallet, decision, action([]string{"cash-a", "cash-b"}[i], "80000", 10), now, cal, fee)
		}(i)
	}
	wg.Wait()
	accepted := 0
	for i, r := range results {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if r.Status == "accepted" {
			accepted++
			if _, e = agent.ApplyAction(ctx, wallet, decision, domain.Action{Action: "CANCEL", ClientID: "cancel-" + r.ActionID, Target: &r.Order.ID}, now, cal, fee); e != nil {
				t.Fatal(e)
			}
		}
	}
	if accepted != 1 {
		t.Fatal("double spend", results)
	}
	x := action("buy", "2000", 10)
	r, e := agent.ApplyAction(ctx, wallet, decision, x, now, cal, fee)
	if e != nil || r.Status != "accepted" {
		t.Fatal(r, e)
	}
	again, e := agent.ApplyAction(ctx, wallet, decision, x, now, cal, fee)
	if e != nil || again.Order.ID != r.Order.ID {
		t.Fatal("idempotence", e)
	}
	changed := x
	changed.Quantity = qty(11)
	if _, e = agent.ApplyAction(ctx, wallet, decision, changed, now, cal, fee); e == nil {
		t.Fatal("different body key accepted")
	}
	snapshot := func(at time.Time, size int64) marketdata.Snapshot {
		rows := []marketdata.Observation{}
		for _, source := range []string{"stockanalysis", "yahoo", "webull"} {
			stamp := at
			rows = append(rows, marketdata.Observation{ID: uuid.NewString(), AssetID: asset, SourceID: source, IdentityVerified: true, Currency: "USD", PriceType: "LAST", Session: "REGULAR", EventTime: &stamp, ObservedAt: at, Price: str("100"), Bid: str("99.99"), Ask: str("100"), BidSize: qty(size), AskSize: qty(size), BBOTime: &stamp, SizeUnit: "shares", VenueScope: "staging-fixture"})
		}
		return marketdata.Snapshot{Asset: marketdata.Asset{ID: asset}, Observations: rows}
	}
	s := snapshot(now.Add(time.Second), 4)
	r, e = worker.Execute(ctx, wallet, r.Order.ID, s, now.Add(time.Second), cal, fee, true)
	if e != nil || r.Fill == nil || r.Fill.Quantity != 4 {
		t.Fatal(r, e)
	}
	oid := r.Order.ID
	var fills int
	if e = owner.Pool.QueryRow(ctx, "SELECT count(*) FROM fills WHERE wallet_id=$1 AND order_id=$2", wallet, oid).Scan(&fills); e != nil || fills != 1 {
		t.Fatal(fills, e)
	}
	repeat, e := worker.Execute(ctx, wallet, oid, s, now.Add(2*time.Second), cal, fee, true)
	if e != nil || repeat.Fill != nil || repeat.Error != "liquidity_consumed" {
		t.Fatal(repeat, e)
	}
	// A fill/cancel race leaves a coherent order and journal, whichever locks first.
	s = snapshot(now.Add(3*time.Second), 20)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := worker.Execute(ctx, wallet, oid, s, now.Add(3*time.Second), cal, fee, true)
		if e != nil {
			t.Error(e)
		}
	}()
	go func() {
		defer wg.Done()
		_, e := agent.ApplyAction(ctx, wallet, decision, domain.Action{Action: "CANCEL", ClientID: "race-cancel", Target: &oid}, now.Add(3*time.Second), cal, fee)
		if e != nil {
			t.Error(e)
		}
	}()
	wg.Wait()
	if e = worker.Reconcile(ctx, wallet); e != nil {
		t.Fatal(e)
	}
	w, e := worker.Wallet(ctx, wallet)
	if e != nil || w.Account.Check() != nil {
		t.Fatal(e)
	}
	t.Log("parallel cash, action idempotence/conflict, partial fill, quote replay, fill/cancel race, journal reconciliation passed; staging fixtures only")
}
func TestPostgresOCOAndFillRollback(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); e != nil {
		t.Fatal(e)
	}
	owner, e := Open(ctx, os.Getenv("TEST_OWNER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Close()
	worker, e := Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	agent, e := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer agent.Close()
	cal, _ := calendar.Load("../../config/calendar.json")
	fee, e := domain.LoadFees("../../config/fees.json")
	if e != nil {
		t.Fatal(e)
	}
	now := time.Date(2026, 10, 12, 15, 0, 0, time.Now().Nanosecond(), time.UTC)
	wallet := WalletID(Codes[1]).String()
	asset := uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:asset:AAPL")).String()
	did := uuid.NewString()
	_, e = owner.Pool.Exec(ctx, "INSERT INTO decision_runs(id,wallet_id,state,input,model,reasoning,started_at) VALUES($1,$2,'ready','{\"evidence_ids\":[]}','fixture','low',$3)", did, wallet, now)
	if e != nil {
		t.Fatal(e)
	}
	x := domain.Action{Action: "SUBMIT", ClientID: "buy", AssetID: &asset, Side: str("BUY"), Quantity: qty(10), Type: str("MARKET"), TIF: str("DAY"), Budget: str("2000")}
	r, e := agent.ApplyAction(ctx, wallet, did, x, now, cal, fee)
	if e != nil || r.Order == nil {
		t.Fatal(r, e)
	}
	snapshot := func(at time.Time, last string, size int64) marketdata.Snapshot {
		out := marketdata.Snapshot{Asset: marketdata.Asset{ID: asset}}
		for _, source := range []string{"stockanalysis", "yahoo", "webull"} {
			stamp := at
			out.Observations = append(out.Observations, marketdata.Observation{ID: uuid.NewString(), AssetID: asset, SourceID: source, IdentityVerified: true, Currency: "USD", PriceType: "LAST", Session: "REGULAR", Price: &last, Bid: &last, Ask: &last, EventTime: &stamp, ObservedAt: at, BBOTime: &stamp, BidSize: qty(size), AskSize: qty(size), SizeUnit: "shares", VenueScope: "staging-fixture"})
		}
		return out
	}
	// Failure late in the transaction must remove the fill, lot, fee, cash change and consumption together.
	_, e = owner.Pool.Exec(ctx, `CREATE OR REPLACE FUNCTION reject_staging_receipt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.body ? 'fill' THEN RAISE EXCEPTION 'injected_receipt_failure'; END IF; RETURN NEW; END $$`)
	if e != nil {
		t.Fatal(e)
	}
	_, e = owner.Pool.Exec(ctx, "CREATE TRIGGER staging_receipt_failure BEFORE INSERT ON receipts FOR EACH ROW EXECUTE FUNCTION reject_staging_receipt()")
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Pool.Exec(ctx, "DROP TRIGGER IF EXISTS staging_receipt_failure ON receipts; DROP FUNCTION IF EXISTS reject_staging_receipt()")
	before, _ := worker.Wallet(ctx, wallet)
	_, e = worker.Execute(ctx, wallet, r.Order.ID, snapshot(now.Add(time.Second), "100", 10), now.Add(time.Second), cal, fee, true)
	if e == nil {
		t.Fatal("expected injected rollback")
	}
	after, _ := worker.Wallet(ctx, wallet)
	if testHash(testMarshal(before.Account)) != testHash(testMarshal(after.Account)) {
		t.Fatal("partial transaction persisted")
	}
	owner.Pool.Exec(ctx, "DROP TRIGGER staging_receipt_failure ON receipts; DROP FUNCTION reject_staging_receipt()")
	r, e = worker.Execute(ctx, wallet, r.Order.ID, snapshot(now.Add(time.Second), "100", 10), now.Add(time.Second), cal, fee, true)
	if e != nil || r.Fill == nil {
		t.Fatal(r, e)
	}
	group := uuid.NewString()
	limit := domain.Action{Action: "SUBMIT", ClientID: "profit", AssetID: &asset, Side: str("SELL"), Quantity: qty(10), Type: str("LIMIT"), TIF: str("DAY"), Limit: str("105"), OCO: &group}
	stop := limit
	stop.ClientID = "stop"
	stop.Type = str("STOP_MARKET")
	stop.Limit = nil
	stop.Stop = str("95")
	l, e := agent.ApplyAction(ctx, wallet, did, limit, now.Add(2*time.Second), cal, fee)
	if e != nil || l.Order == nil {
		t.Fatal(l, e)
	}
	sl, e := agent.ApplyAction(ctx, wallet, did, stop, now.Add(2*time.Second), cal, fee)
	if e != nil || sl.Order == nil {
		t.Fatal(sl, e)
	}
	var reserved int64
	if e = owner.Pool.QueryRow(ctx, "SELECT sum(shares)::bigint FROM reservations WHERE wallet_id=$1 AND oco_id=$2", wallet, group).Scan(&reserved); e != nil || reserved != 10 {
		t.Fatal("double OCO reserve", reserved, e)
	}
	// Both workers see sufficient quantity; wallet lock and OCO reduction prevent overselling.
	var wg sync.WaitGroup
	for _, order := range []string{l.Order.ID, sl.Order.ID} {
		wg.Add(1)
		go func(order string) {
			defer wg.Done()
			price := "106"
			if order == sl.Order.ID {
				price = "90"
			}
			_, e := worker.Execute(ctx, wallet, order, snapshot(now.Add(3*time.Second), price, 10), now.Add(3*time.Second), cal, fee, true)
			if e != nil {
				t.Error(e)
			}
		}(order)
	}
	wg.Wait()
	after, e = worker.Wallet(ctx, wallet)
	if e != nil || after.Account.Shares(asset) != before.Account.Shares(asset) {
		t.Fatal("OCO oversell", e, after.Account)
	}
	if e = worker.Reconcile(ctx, wallet); e != nil {
		t.Fatal(e)
	}
	due, _ := cal.Settlement(now)
	if e = worker.Settle(ctx, wallet, due); e != nil {
		t.Fatal(e)
	}
	if e = worker.Settle(ctx, wallet, due); e != nil {
		t.Fatal(e)
	}
	if e = worker.Reconcile(ctx, wallet); e != nil {
		t.Fatal(e)
	}
	t.Log("late transaction rollback, one OCO reservation, concurrent OCO legs, settlement replay and reconciliation passed")
}
func TestPostgresCorporateCash(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); e != nil {
		t.Fatal(e)
	}
	owner, e := Open(ctx, os.Getenv("TEST_OWNER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Close()
	worker, e := Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	agent, e := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer agent.Close()
	cal, _ := calendar.Load("../../config/calendar.json")
	fee, _ := domain.LoadFees("../../config/fees.json")
	wallet := WalletID(Codes[2]).String()
	asset := uuid.NewString()
	now := time.Date(2026, 10, 12, 16, 0, 0, time.Now().Nanosecond(), time.UTC)
	did := uuid.NewString()
	// Unique staging fixture avoids mixing chronological corporate histories on repeat tests.
	_, e = owner.Pool.Exec(ctx, "INSERT INTO assets(id,symbol,name,category,exchange,aliases) VALUES($1,$2,'Staging corporate fixture','US','NASDAQ','{}')", asset, "TEST-"+asset)
	if e != nil {
		t.Fatal(e)
	}
	_, e = owner.Pool.Exec(ctx, "INSERT INTO decision_runs(id,wallet_id,state,input,model,reasoning,started_at) VALUES($1,$2,'ready','{\"evidence_ids\":[]}','fixture','low',$3)", did, wallet, now)
	if e != nil {
		t.Fatal(e)
	}
	x := domain.Action{Action: "SUBMIT", ClientID: "corp-buy", AssetID: &asset, Side: str("BUY"), Quantity: qty(10), Type: str("MARKET"), TIF: str("DAY"), Budget: str("2000")}
	r, e := agent.ApplyAction(ctx, wallet, did, x, now, cal, fee)
	if e != nil || r.Order == nil {
		t.Fatal(r, e)
	}
	at := now.Add(time.Second)
	snapshot := marketdata.Snapshot{Asset: marketdata.Asset{ID: asset}}
	for _, source := range []string{"stockanalysis", "yahoo", "webull"} {
		snapshot.Observations = append(snapshot.Observations, marketdata.Observation{ID: uuid.NewString(), AssetID: asset, SourceID: source, IdentityVerified: true, Currency: "USD", PriceType: "LAST", Session: "REGULAR", EventTime: &at, ObservedAt: at, Price: str("100"), Bid: str("100"), Ask: str("100"), BidSize: qty(10), AskSize: qty(10), BBOTime: &at, SizeUnit: "shares", VenueScope: "staging-fixture"})
	}
	if r, e = worker.Execute(ctx, wallet, r.Order.ID, snapshot, at, cal, fee, true); e != nil || r.Fill == nil {
		t.Fatal(r, e)
	}
	split := domain.CorporateEvent{ID: uuid.NewString(), AssetID: asset, Kind: "SPLIT", EffectiveAt: now.Add(time.Minute), Verified: true, Ratio: "2", SourceURL: "https://www.sec.gov/staging-fixture"}
	insert := func(ev domain.CorporateEvent) {
		_, e := owner.Pool.Exec(ctx, "INSERT INTO corporate_events(id,asset_id,event_type,effective_at,status,source_url,body) VALUES($1,$2,$3,$4,'verified',$5,$6)", ev.ID, asset, ev.Kind, ev.EffectiveAt, ev.SourceURL, testMarshal(ev))
		if e != nil {
			t.Fatal(e)
		}
	}
	insert(split)
	for range 2 {
		if e = worker.Corporate(ctx, wallet, split.ID, split.EffectiveAt); e != nil {
			t.Fatal(e)
		}
	}
	w, e := worker.Wallet(ctx, wallet)
	if e != nil || w.Account.Shares(asset) != 20 {
		t.Fatal(w, e)
	}
	div := domain.CorporateEvent{ID: uuid.NewString(), AssetID: asset, Kind: "DIVIDEND", EffectiveAt: now.Add(2 * time.Minute), PayAt: now.AddDate(0, 0, 2), Verified: true, PerShare: "1", Classification: "ordinary_us", SourceURL: "https://www.sec.gov/staging-fixture"}
	insert(div)
	for range 2 {
		if e = worker.Corporate(ctx, wallet, div.ID, div.EffectiveAt); e != nil {
			t.Fatal(e)
		}
	}
	w, e = worker.Wallet(ctx, wallet)
	if e != nil || w.Account.Dividend != "14" {
		t.Fatal(w, e)
	}
	cash := w.Account.Available
	for range 2 {
		if e = worker.Settle(ctx, wallet, div.PayAt); e != nil {
			t.Fatal(e)
		}
	}
	w, e = worker.Wallet(ctx, wallet)
	if e != nil || w.Account.Dividend != "0" || w.Account.Available == cash {
		t.Fatal(w, e)
	}
	if e = worker.Reconcile(ctx, wallet); e != nil {
		t.Fatal(e)
	}
	t.Log("split, historical ex-date entitlement, 30% modeled withholding, dividend payment and event replay passed; unique staging fixture")
}
func TestScopedActionFencing(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); e != nil {
		t.Fatal(e)
	}
	owner, e := Open(ctx, os.Getenv("TEST_OWNER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer owner.Close()
	worker, e := Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	agent, e := Open(ctx, os.Getenv("TEST_AGENT_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer agent.Close()
	q := jobs.Queue{Pool: worker.Pool}
	wid := WalletID(Codes[7])
	kind := "fence-" + uuid.NewString()
	id, e := q.Enqueue(ctx, jobs.Job{Kind: kind, Key: uuid.NewString(), Wallet: &wid, Available: time.Now().Add(-time.Second), Deadline: time.Now().Add(time.Minute)}, nil, 0)
	if e != nil {
		t.Fatal(e)
	}
	j, e := q.Acquire(ctx, "scope-fixture", []string{kind})
	if e != nil || j == nil {
		t.Fatal(j, e)
	}
	did := uuid.NewString()
	now := time.Now().UTC()
	_, e = owner.Pool.Exec(ctx, "INSERT INTO decision_runs(id,wallet_id,job_id,fencing_token,state,input,model,reasoning,started_at) VALUES($1,$2,$3,$4,'ready','{\"evidence_ids\":[]}','fixture','low',$5)", did, wid, id, j.Fence, now)
	if e != nil {
		t.Fatal(e)
	}
	cal, _ := calendar.Load("../../config/calendar.json")
	fee, _ := domain.LoadFees("../../config/fees.json")
	x := domain.Action{Action: "HOLD", ClientID: "fenced-hold", Hold: str("Fixture")}
	if r, e := agent.ApplyAction(ctx, wid.String(), did, x, now, cal, fee, ActionScope{Job: id.String(), Fence: j.Fence}); e != nil || r.Status != "held" {
		t.Fatal("valid scoped lock", r, e)
	}
	_, e = worker.Pool.Exec(ctx, "UPDATE jobs SET lease_until=now()-interval '1 second' WHERE id=$1", id)
	if e != nil {
		t.Fatal(e)
	}
	newJob, e := q.Acquire(ctx, "reclaimed", []string{kind})
	if e != nil || newJob == nil {
		t.Fatal(e)
	}
	x.ClientID = "stale-hold"
	if _, e = agent.ApplyAction(ctx, wid.String(), did, x, now, cal, fee, ActionScope{Job: id.String(), Fence: j.Fence}); e == nil {
		t.Fatal("stale action mutated")
	}
	if e = q.Finish(ctx, *newJob, nil, nil); e != nil {
		t.Fatal(e)
	}
	t.Log("scoped agent can lock only own job; stale action rejected after lease reclamation")
}

func testMarshal(v any) []byte { b, _ := json.Marshal(v); return b }
func testHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
