package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"regexp"
	"time"
)

var actionID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func marshal(v any) []byte {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func clone(a domain.Account) domain.Account {
	var out domain.Account
	json.Unmarshal(marshal(a), &out)
	return out
}
func balances(a domain.Account) map[string]decimal.Decimal {
	d := func(v string) decimal.Decimal {
		if v == "" {
			return decimal.Zero
		}
		x, _ := domain.Money(v)
		return x
	}
	cost := decimal.Zero
	for _, l := range a.Lots {
		cost = cost.Add(d(l.Cost))
	}
	return map[string]decimal.Decimal{"cash_available": d(a.Available), "cash_reserved": d(a.Reserved), "cash_unsettled": d(a.Unsettled), "dividend_receivable": d(a.Dividend), "position_cost": cost, "dividend_income": d(a.DividendIncome).Neg(), "realized_income": d(a.Realized).Neg()}
}
func persist(ctx context.Context, tx pgx.Tx, q *dbgen.Queries, id uuid.UUID, before, after domain.Account, event, key string, now time.Time) error {
	if e := after.Check(); e != nil {
		return e
	}
	if _, e := tx.Exec(ctx, "INSERT INTO account_events(id,wallet_id,event_type,external_event_id,occurred_at,state_hash) VALUES($1,$2,$3,$4,$5,$6)", uuid.New(), id, event, key, now, hash(marshal(after))); e != nil {
		return e
	}
	old, updated := balances(before), balances(after)
	deltas := map[string]decimal.Decimal{}
	total := decimal.Zero
	for account, v := range updated {
		diff := v.Sub(old[account])
		if !diff.IsZero() {
			deltas[account] = diff
			total = total.Add(diff)
		}
	}
	if !total.IsZero() {
		return fmt.Errorf("journal_state_not_balanced")
	}
	if len(deltas) > 0 {
		jid := uuid.New()
		_, e := tx.Exec(ctx, "INSERT INTO journal_transactions(id,wallet_id,event_type,external_event_id,occurred_at) VALUES($1,$2,$3,$4,$5)", jid, id, event, key, now)
		if e != nil {
			return e
		}
		for account, amount := range deltas {
			if _, e = tx.Exec(ctx, "INSERT INTO journal_lines(wallet_id,transaction_id,account_code,amount,currency) VALUES($1,$2,$3,$4::numeric,'USD')", id, jid, account, amount.String()); e != nil {
				return e
			}
		}
	}
	groupOwners := map[string]string{}
	groupQty := map[string]int64{}
	groupAsset := map[string]string{}
	for _, o := range after.Orders {
		if o.OCO != nil {
			g := *o.OCO
			groupAsset[g] = o.AssetID
			if domain.Active(o) {
				if groupOwners[g] == "" || o.ID < groupOwners[g] {
					groupOwners[g] = o.ID
				}
				if o.Remaining > groupQty[g] {
					groupQty[g] = o.Remaining
				}
			}
		}
	}
	for g, asset := range groupAsset {
		if _, e := tx.Exec(ctx, "INSERT INTO oco_groups(wallet_id,id,asset_id,reserved_shares) VALUES($1,$2,$3,$4) ON CONFLICT(wallet_id,id) DO UPDATE SET reserved_shares=excluded.reserved_shares", id, g, asset, groupQty[g]); e != nil {
			return e
		}
	}
	for _, o := range after.Orders {
		_, e := tx.Exec(ctx, "INSERT INTO orders(id,wallet_id,asset_id,state,body,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(id) DO UPDATE SET state=excluded.state,body=excluded.body,updated_at=excluded.updated_at WHERE orders.wallet_id=excluded.wallet_id", o.ID, id, o.AssetID, o.State, marshal(o), o.CreatedAt, now)
		if e != nil {
			return e
		}
		if before.Orders[o.ID] == nil || hash(marshal(before.Orders[o.ID])) != hash(marshal(o)) {
			if _, e = tx.Exec(ctx, "INSERT INTO order_events(wallet_id,order_id,body,occurred_at) VALUES($1,$2,$3,$4)", id, o.ID, marshal(o), now); e != nil {
				return e
			}
		}
		shares := int64(0)
		if o.Side == "SELL" && domain.Active(o) {
			shares = o.Remaining
			if o.OCO != nil {
				shares = 0
				if groupOwners[*o.OCO] == o.ID {
					shares = groupQty[*o.OCO]
				}
			}
		}
		_, e = tx.Exec(ctx, "INSERT INTO reservations(wallet_id,order_id,oco_id,cash_usd,shares) VALUES($1,$2,$3,$4::numeric,$5) ON CONFLICT(wallet_id,order_id) DO UPDATE SET oco_id=excluded.oco_id,cash_usd=excluded.cash_usd,shares=excluded.shares", id, o.ID, o.OCO, o.ReservedCash, shares)
		if e != nil {
			return e
		}
	}
	for _, l := range after.Lots {
		_, e := tx.Exec(ctx, "INSERT INTO position_lots(id,wallet_id,asset_id,fill_id,quantity,cost_basis,acquired_at) VALUES($1,$2,$3,$4,$5,$6::numeric,$7) ON CONFLICT(id) DO UPDATE SET quantity=excluded.quantity,cost_basis=excluded.cost_basis WHERE position_lots.wallet_id=excluded.wallet_id", l.ID, id, l.AssetID, l.FillID, l.Quantity, l.Cost, l.AcquiredAt)
		if e != nil {
			return e
		}
	}
	for _, settlement := range after.Settlements {
		if _, e := tx.Exec(ctx, "INSERT INTO settlements(wallet_id,fill_id,amount,due_at,settled) VALUES($1,$2,$3::numeric,$4,$5) ON CONFLICT(wallet_id,fill_id) DO UPDATE SET settled=excluded.settled", id, settlement.ID, settlement.Amount, settlement.Due, settlement.Done); e != nil {
			return e
		}
	}
	return q.UpdateWallet(ctx, dbgen.UpdateWalletParams{ID: id, State: marshal(after)})
}
func (s *Store) ApplyAction(ctx context.Context, wallet, decision string, x domain.Action, now time.Time, cal calendar.Calendar, fee domain.FeeProfile, scopes ...ActionScope) (domain.Receipt, error) {
	var out domain.Receipt
	for _, id := range []*string{x.AssetID, x.Target, x.OCO} {
		if id != nil {
			if _, e := uuid.Parse(*id); e != nil {
				return out, fmt.Errorf("invalid_identifier")
			}
		}
	}
	if !actionID.MatchString(x.ClientID) {
		return out, fmt.Errorf("invalid_client_action_id")
	}
	did, e := uuid.Parse(decision)
	if e != nil {
		return out, fmt.Errorf("invalid_decision")
	}
	bodyHash := hash(marshal(x))
	e = s.Scoped(ctx, wallet, func(tx pgx.Tx, q *dbgen.Queries) error {
		wid := uuid.MustParse(wallet)
		row, e := q.LockWallet(ctx, wid)
		if e != nil {
			return e
		}
		existing, e := q.GetAction(ctx, dbgen.GetActionParams{WalletID: wid, DecisionID: did, ClientActionID: x.ClientID})
		if e == nil {
			if existing.BodyHash != bodyHash {
				return fmt.Errorf("idempotency_conflict")
			}
			return json.Unmarshal(existing.Receipt, &out)
		}
		if e != pgx.ErrNoRows {
			return e
		}
		if len(scopes) > 0 {
			scope := scopes[0]
			var jobID uuid.UUID
			var fence int64
			var owner string
			e = tx.QueryRow(ctx, "SELECT d.job_id,d.fencing_token,j.lease_owner FROM decision_runs d JOIN jobs j ON j.id=d.job_id WHERE d.wallet_id=$1 AND d.id=$2", wid, did).Scan(&jobID, &fence, &owner)
			if e != nil {
				return fmt.Errorf("decision_scope_missing")
			}
			if jobID.String() != scope.Job || fence != scope.Fence {
				return fmt.Errorf("decision_scope_mismatch")
			}
			if e = jobs.Guard(ctx, tx, jobs.Job{ID: jobID, Owner: owner, Fence: fence}); e != nil {
				return e
			}
		}
		var input []byte
		var state string
		var start time.Time
		e = tx.QueryRow(ctx, "SELECT state,input,started_at FROM decision_runs WHERE wallet_id=$1 AND id=$2", wid, did).Scan(&state, &input, &start)
		if e != nil {
			return fmt.Errorf("decision_not_found")
		}
		if (state != "ready" && state != "running") || now.Sub(start) > 10*time.Minute || start.After(now) {
			return fmt.Errorf("decision_not_actionable")
		}
		var a domain.Account
		if e = json.Unmarshal(row.State, &a); e != nil {
			return e
		}
		if e = a.Check(); e != nil {
			return e
		}
		before := clone(a)
		out = domain.Receipt{ActionID: x.ClientID, Status: "rejected", AsOf: now}
		var opErr error
		for _, dependency := range x.Depends {
			var r []byte
			e = tx.QueryRow(ctx, "SELECT receipt FROM decision_actions WHERE wallet_id=$1 AND decision_id=$2 AND client_action_id=$3", wid, did, dependency).Scan(&r)
			var receipt domain.Receipt
			if e != nil || json.Unmarshal(r, &receipt) != nil || receipt.Status == "rejected" {
				opErr = fmt.Errorf("dependent_action_blocked")
				break
			}
		}
		if opErr == nil {
			var manifest struct {
				EvidenceIDs []string `json:"evidence_ids"`
			}
			json.Unmarshal(input, &manifest)
			seen := map[string]bool{}
			for _, id := range manifest.EvidenceIDs {
				seen[id] = true
			}
			for _, id := range x.EvidenceIDs {
				if !seen[id] {
					var found bool
					if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM decision_evidence WHERE wallet_id=$1 AND decision_id=$2 AND evidence_id=$3)", wid, did, id).Scan(&found); e != nil {
						return e
					}
					if found {
						continue
					}
					opErr = fmt.Errorf("evidence_not_seen")
					break
				}
			}
		}
		if opErr == nil {
			switch x.Action {
			case "HOLD":
				if x.Hold == nil || *x.Hold == "" {
					opErr = fmt.Errorf("hold_reason_required")
				} else {
					out.Status = "held"
				}
			case "CANCEL":
				if x.Target == nil {
					opErr = fmt.Errorf("target_order_required")
				} else {
					opErr = a.Cancel(*x.Target)
					if opErr == nil {
						out.Order = a.Orders[*x.Target]
						out.Status = "cancelled"
					}
				}
			case "SUBMIT", "REPLACE":
				allowed := false
				if x.AssetID != nil {
					var category string
					var fixed bool
					e = tx.QueryRow(ctx, "SELECT category,(fixed OR id IN (SELECT asset_id FROM dynamic_members WHERE version_id=(SELECT id FROM universe_versions ORDER BY effective_at DESC LIMIT 1)) OR id IN (SELECT asset_id FROM wallet_watchlist_members WHERE wallet_id=$3 AND active) OR ($2 AND EXISTS(SELECT 1 FROM position_lots WHERE asset_id=assets.id AND wallet_id=$3 AND quantity>0))) FROM assets WHERE id=$1", *x.AssetID, x.Side != nil && *x.Side == "SELL", wid).Scan(&category, &fixed)
					if e == nil {
						allowed = category == row.Category && (row.Category != "US" || fixed)
					}
				}
				if x.Action == "REPLACE" {
					if x.Target == nil {
						opErr = fmt.Errorf("target_order_required")
					} else {
						opErr = a.Cancel(*x.Target)
					}
				}
				if opErr == nil {
					oid := uuid.NewSHA1(did, []byte(wallet+":"+x.ClientID)).String()
					opErr = a.Submit(x, oid, now, cal, fee, allowed)
					if opErr == nil {
						out.Order = a.Orders[oid]
						out.Status = "accepted"
					}
				}
			default:
				opErr = fmt.Errorf("invalid_action")
			}
		}
		if opErr != nil {
			a = before
			out.Error = opErr.Error()
		} else {
			if e = persist(ctx, tx, q, wid, before, a, "order_action", did.String()+":"+x.ClientID, now); e != nil {
				return e
			}
		}
		out.Account = &a
		b := marshal(out)
		_, e = tx.Exec(ctx, "INSERT INTO decision_actions(wallet_id,decision_id,client_action_id,body_hash,receipt) VALUES($1,$2,$3,$4,$5)", wid, did, x.ClientID, bodyHash, b)
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO receipts(wallet_id,decision_id,body) VALUES($1,$2,$3)", wid, did, b)
		return e
	})
	return out, e
}
func (s *Store) Execute(ctx context.Context, wallet, order string, snapshot marketdata.Snapshot, now time.Time, cal calendar.Calendar, fee domain.FeeProfile, enabled bool, leases ...jobs.Job) (domain.Receipt, error) {
	out := domain.Receipt{AsOf: now}
	if !enabled {
		out.Status = "waiting"
		out.Error = "execution_disabled"
		return out, nil
	}
	e := s.Scoped(ctx, wallet, func(tx pgx.Tx, q *dbgen.Queries) error {
		wid := uuid.MustParse(wallet)
		row, e := q.LockWallet(ctx, wid)
		if e != nil {
			return e
		}
		for _, j := range leases {
			if e := jobs.Guard(ctx, tx, j); e != nil {
				return e
			}
		}
		var a domain.Account
		if e = json.Unmarshal(row.State, &a); e != nil {
			return e
		}
		before := clone(a)
		var blocked bool
		var epoch *time.Time
		if e = tx.QueryRow(ctx, "SELECT execution_blocked,quote_not_before FROM assets WHERE id=$1", snapshot.Asset.ID).Scan(&blocked, &epoch); e != nil {
			return e
		}
		if blocked {
			out.Status = "waiting"
			out.Error = "corporate_action_unresolved"
			return nil
		}
		if epoch != nil {
			for _, o := range snapshot.Observations {
				if o.ObservedAt.Before(*epoch) {
					out.Status = "waiting"
					out.Error = "post_corporate_quote_required"
					return nil
				}
			}
		}
		fill, opErr := a.TryFill(order, snapshot, now, cal, fee)
		out.Order = a.Orders[order]
		out.Account = &a
		if fill != nil {
			b := marshal(fill)
			_, e = tx.Exec(ctx, "INSERT INTO fills(id,wallet_id,order_id,body,occurred_at) VALUES($1,$2,$3,$4,$5)", fill.ID, wid, order, b, now)
			if e != nil {
				return e
			}
			if snapshot.ExecutionMode == marketdata.ObservedPriceMode {
				_, e = tx.Exec(ctx, "INSERT INTO execution_quotes(id,asset_id,body,observed_at) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING", fill.QuoteID, snapshot.Asset.ID, marshal(snapshot), now)
				if e != nil {
					return e
				}
			}
			for _, ob := range snapshot.Observations {
				if domain.QuoteKey(ob) == fill.QuoteID {
					_, e = tx.Exec(ctx, "INSERT INTO execution_quotes(id,asset_id,body,observed_at) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING", fill.QuoteID, snapshot.Asset.ID, marshal(ob), ob.ObservedAt)
					if e != nil {
						return e
					}
				}
			}
			_, e = tx.Exec(ctx, "INSERT INTO quote_consumptions(wallet_id,quote_id,side,quantity) VALUES($1,$2,$3,$4) ON CONFLICT(wallet_id,quote_id,side) DO UPDATE SET quantity=quote_consumptions.quantity+excluded.quantity", wid, fill.QuoteID, a.Orders[order].Side, fill.Quantity)
			if e != nil {
				return e
			}
			out.Fill = fill
			out.Status = a.Orders[order].State
		} else {
			out.Status = "waiting"
			if opErr != nil {
				out.Error = opErr.Error()
			}
			if out.Order != nil && !domain.Active(out.Order) {
				out.Status = out.Order.State
			}
		}
		if hash(marshal(a)) != hash(marshal(before)) {
			key := order + ":" + now.Format(time.RFC3339Nano)
			if fill != nil {
				key = fill.ID
			}
			if e = persist(ctx, tx, q, wid, before, a, "execution", key, now); e != nil {
				return e
			}
			_, e = tx.Exec(ctx, "INSERT INTO receipts(wallet_id,body) VALUES($1,$2)", wid, marshal(out))
			if e != nil {
				return e
			}
		}
		return nil
	})
	return out, e
}
func (s *Store) Settle(ctx context.Context, wallet string, now time.Time, leases ...jobs.Job) error {
	return s.Scoped(ctx, wallet, func(tx pgx.Tx, q *dbgen.Queries) error {
		wid := uuid.MustParse(wallet)
		r, e := q.LockWallet(ctx, wid)
		if e != nil {
			return e
		}
		for _, j := range leases {
			if e := jobs.Guard(ctx, tx, j); e != nil {
				return e
			}
		}
		var a domain.Account
		if e = json.Unmarshal(r.State, &a); e != nil {
			return e
		}
		before := clone(a)
		a.Settle(now)
		a.Expire(now)
		a.PayDividends(now)
		for _, r := range a.Receivables {
			if _, e = tx.Exec(ctx, "UPDATE receivables SET status=$3 WHERE wallet_id=$1 AND event_id=$2", wid, r.ID, r.Status); e != nil {
				return e
			}
		}
		if hash(marshal(a)) == hash(marshal(before)) {
			return nil
		}
		return persist(ctx, tx, q, wid, before, a, "settlement", fmt.Sprintf("%d:%s", r.AccountVersion, now.Format(time.RFC3339Nano)), now)
	})
}

// Reconcile compares the account view against immutable journal and FIFO rows.
func (s *Store) Reconcile(ctx context.Context, wallet string) error {
	return s.Scoped(ctx, wallet, func(tx pgx.Tx, q *dbgen.Queries) error {
		wid := uuid.MustParse(wallet)
		row, e := q.GetWallet(ctx, wid)
		if e != nil {
			return e
		}
		var a domain.Account
		if e = json.Unmarshal(row.State, &a); e != nil {
			return e
		}
		if e = a.Check(); e != nil {
			return e
		}
		expected := balances(a)
		rows, e := tx.Query(ctx, "SELECT account_code,sum(amount)::text FROM journal_lines WHERE wallet_id=$1 GROUP BY account_code", wid)
		if e != nil {
			return e
		}
		defer rows.Close()
		seen := map[string]bool{}
		for rows.Next() {
			var code, amount string
			if e = rows.Scan(&code, &amount); e != nil {
				return e
			}
			if code == "capital" {
				continue
			}
			d, e := domain.Money(amount)
			if e != nil || !d.Equal(expected[code]) {
				return fmt.Errorf("journal_reconciliation_failed: %s", code)
			}
			seen[code] = true
		}
		if e = rows.Err(); e != nil {
			return e
		}
		for code, d := range expected {
			if !d.IsZero() && !seen[code] {
				return fmt.Errorf("journal_account_missing: %s", code)
			}
		}

		positionRows, e := tx.Query(ctx, "SELECT asset_id::text,sum(quantity)::bigint,sum(cost_basis)::text FROM position_lots WHERE wallet_id=$1 GROUP BY asset_id", wid)
		if e != nil {
			return e
		}
		defer positionRows.Close()
		basisTotal := decimal.Zero
		for positionRows.Next() {
			var asset, cost string
			var qty int64
			if e = positionRows.Scan(&asset, &qty, &cost); e != nil {
				return e
			}
			if qty != a.Shares(asset) {
				return fmt.Errorf("position_reconciliation_failed")
			}
			d, e := domain.Money(cost)
			if e != nil {
				return e
			}
			basisTotal = basisTotal.Add(d)
		}
		if e = positionRows.Err(); e != nil {
			return e
		}
		if !basisTotal.Equal(expected["position_cost"]) {
			return fmt.Errorf("cost_basis_reconciliation_failed")
		}
		return nil
	})
}
func (s *Store) Corporate(ctx context.Context, wallet, eventID string, now time.Time, leases ...jobs.Job) error {
	return s.Scoped(ctx, wallet, func(tx pgx.Tx, q *dbgen.Queries) error {
		wid := uuid.MustParse(wallet)
		row, e := q.LockWallet(ctx, wid)
		if e != nil {
			return e
		}
		for _, j := range leases {
			if e := jobs.Guard(ctx, tx, j); e != nil {
				return e
			}
		}
		var a domain.Account
		if e = json.Unmarshal(row.State, &a); e != nil {
			return e
		}
		var ev domain.CorporateEvent
		var body []byte
		var status, asset string
		var correction *string
		if e = tx.QueryRow(ctx, "SELECT body,status,asset_id::text,supersedes::text FROM corporate_events WHERE id=$1", eventID).Scan(&body, &status, &asset, &correction); e != nil {
			return e
		}
		if e = eventIdentity(body, eventID, asset); e != nil {
			return e
		}
		if correction != nil {
			return fmt.Errorf("corporate_revision_unresolved")
		}
		if e = json.Unmarshal(body, &ev); e != nil {
			return e
		}
		if status != "verified" || !ev.Verified || now.Before(ev.EffectiveAt) {
			return fmt.Errorf("corporate_event_not_actionable")
		}
		if a.AppliedEvents[ev.ID] {
			return nil
		}
		before := clone(a)
		switch ev.Kind {
		case "SPLIT":
			e = a.Split(ev)
		case "DIVIDEND":
			var qty int64
			qty, e = historicalShares(ctx, tx, wallet, ev.AssetID, ev.EffectiveAt)
			if e == nil {
				e = a.DividendEntitlement(ev, qty)
			}
		default:
			e = fmt.Errorf("corporate_action_unresolved")
		}
		if e != nil {
			return e
		}
		if e = persist(ctx, tx, q, wid, before, a, "corporate_event", ev.ID, now); e != nil {
			return e
		}
		for _, r := range a.Receivables {
			if r.ID == ev.ID {
				_, e = tx.Exec(ctx, "INSERT INTO receivables(wallet_id,event_id,gross,tax,net,quantity,pay_at,status) VALUES($1,$2,$3::numeric,$4::numeric,$5::numeric,$6,$7,$8) ON CONFLICT(wallet_id,event_id) DO NOTHING", wid, r.ID, r.Gross, r.Tax, r.Net, r.Quantity, r.PayAt, r.Status)
			}
		}
		return e
	})
}

type ActionScope struct {
	Job   string
	Fence int64
}
