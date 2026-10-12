package store

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"sort"
	"time"
)

type PositionView struct {
	Asset    string     `json:"asset_id"`
	Symbol   string     `json:"symbol"`
	Quantity int64      `json:"quantity"`
	Cost     string     `json:"cost_usd"`
	Price    *string    `json:"price_usd"`
	Value    *string    `json:"value_usd"`
	AsOf     *time.Time `json:"price_as_of"`
	Quality  string     `json:"quality"`
}
type WalletView struct {
	Unrealized    *string           `json:"unrealized_usd"`
	ID            uuid.UUID         `json:"id"`
	Code          string            `json:"code"`
	Category      string            `json:"category"`
	Version       int64             `json:"version"`
	Available     string            `json:"available_usd"`
	Reserved      string            `json:"reserved_usd"`
	Unsettled     string            `json:"unsettled_usd"`
	Dividend      string            `json:"dividend_usd"`
	Fees          string            `json:"fees_usd"`
	Realized      string            `json:"realized_usd"`
	Equity        *string           `json:"equity_usd"`
	PositionValue *string           `json:"position_value_usd"`
	Complete      bool              `json:"valuation_complete"`
	OpenOrders    int               `json:"open_orders"`
	Positions     []PositionView    `json:"positions"`
	Orders        []*domain.Order   `json:"orders"`
	Decisions     []json.RawMessage `json:"decisions"`
	Fills         []json.RawMessage `json:"fills"`
}
type ObservationView struct {
	Assets          map[string]string `json:"assets"`
	AsOf            time.Time         `json:"as_of"`
	Wallets         []WalletView      `json:"wallets"`
	Available       string            `json:"available_usd"`
	Equity          *string           `json:"equity_usd"`
	PositionValue   *string           `json:"position_value_usd"`
	Complete        bool              `json:"valuation_complete"`
	Paused          bool              `json:"paused"`
	Execution       bool              `json:"execution_enabled"`
	LiveAcceptance  json.RawMessage   `json:"live_acceptance"`
	Jobs            []json.RawMessage `json:"jobs"`
	Events          []json.RawMessage `json:"events"`
	PendingReceipts int64             `json:"pending_receipts"`
	CalendarSource  string            `json:"calendar_source"`
}

func ptrMoney(v decimal.Decimal) *string { s := v.String(); return &s }
func val(s string) decimal.Decimal       { v, _ := decimal.NewFromString(s); return v }
func valueWallet(w Wallet, names map[string]string, quotes map[string]marketdata.Snapshot, now time.Time) WalletView {
	a := w.Account
	v := WalletView{ID: w.ID, Code: w.Code, Category: w.Category, Version: w.Version, Available: a.Available, Reserved: a.Reserved, Unsettled: a.Unsettled, Dividend: a.Dividend, Fees: a.Fees, Realized: a.Realized, Complete: true, Positions: []PositionView{}, Orders: []*domain.Order{}, Decisions: []json.RawMessage{}, Fills: []json.RawMessage{}}
	lots := map[string]*PositionView{}
	for _, l := range a.Lots {
		if l.Quantity == 0 {
			continue
		}
		p := lots[l.AssetID]
		if p == nil {
			p = &PositionView{Asset: l.AssetID, Symbol: names[l.AssetID], Cost: "0", Quality: "missing"}
			lots[l.AssetID] = p
		}
		p.Quantity += l.Quantity
		p.Cost = val(p.Cost).Add(val(l.Cost)).String()
	}
	total := decimal.Zero
	for id, p := range lots {
		if q, ok := quotes[id]; ok {
			t := q.Validation.AsOf
			p.AsOf = &t
			if q.Validation.Verified && q.Validation.Median != nil {
				p.Price = q.Validation.Median
				p.Quality = "verified"
				if now.Sub(t) > 60*time.Second || t.After(now) {
					p.Quality = "stale"
				}
			}
			if p.Price == nil {
				for _, o := range q.Observations {
					if o.IdentityVerified && o.Currency == "USD" && o.Price != nil {
						p.Price = o.Price
						p.AsOf = o.EventTime
						p.Quality = "unverified"
						break
					}
				}
			}
		}
		if p.Price != nil {
			d, e := decimal.NewFromString(*p.Price)
			if e == nil && d.IsPositive() {
				n := d.Mul(decimal.NewFromInt(p.Quantity))
				p.Value = ptrMoney(n)
				total = total.Add(n)
			}
		}
		if p.Value == nil {
			v.Complete = false
		}
		v.Positions = append(v.Positions, *p)
	}
	sort.Slice(v.Positions, func(i, j int) bool { return v.Positions[i].Symbol < v.Positions[j].Symbol })
	for _, o := range a.Orders {
		v.Orders = append(v.Orders, o)
		if domain.Active(o) {
			v.OpenOrders++
		}
	}
	sort.Slice(v.Orders, func(i, j int) bool { return v.Orders[i].CreatedAt.After(v.Orders[j].CreatedAt) })
	if v.Complete {
		cost := decimal.Zero
		for _, p := range v.Positions {
			cost = cost.Add(val(p.Cost))
		}
		v.Unrealized = ptrMoney(total.Sub(cost))
		v.PositionValue = ptrMoney(total)
		v.Equity = ptrMoney(total.Add(val(a.Available)).Add(val(a.Reserved)).Add(val(a.Unsettled)).Add(val(a.Dividend)))
	}
	return v
}

// Snapshot uses a read-only repeatable snapshot; no model calls or order writes.
func (s *Store) Observation(ctx context.Context) (ObservationView, error) {
	out := ObservationView{AsOf: time.Now().UTC(), Wallets: []WalletView{}, Jobs: []json.RawMessage{}, Events: []json.RawMessage{}, Complete: true, CalendarSource: "config/calendar.json"}
	tx, e := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, "SELECT id::text,symbol FROM assets")
	if e != nil {
		return out, e
	}
	names := map[string]string{}
	for rows.Next() {
		var id, name string
		if e = rows.Scan(&id, &name); e != nil {
			rows.Close()
			return out, e
		}
		names[id] = name
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	out.Assets = names
	quotes := map[string]marketdata.Snapshot{}
	rows, e = tx.Query(ctx, "SELECT DISTINCT ON(asset_id) asset_id::text,body FROM market_snapshots ORDER BY asset_id,captured_at DESC")
	if e != nil {
		return out, e
	}
	for rows.Next() {
		var id string
		var b []byte
		if e = rows.Scan(&id, &b); e != nil {
			rows.Close()
			return out, e
		}
		var q marketdata.Snapshot
		if e = json.Unmarshal(b, &q); e != nil {
			rows.Close()
			return out, e
		}
		quotes[id] = q
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	rows, e = tx.Query(ctx, "SELECT id,code,category,account_version,state FROM wallets ORDER BY code")
	if e != nil {
		return out, e
	}
	wallets := []Wallet{}
	for rows.Next() {
		var w Wallet
		var b []byte
		if e = rows.Scan(&w.ID, &w.Code, &w.Category, &w.Version, &b); e != nil {
			rows.Close()
			return out, e
		}
		if e = json.Unmarshal(b, &w.Account); e != nil {
			rows.Close()
			return out, e
		}
		wallets = append(wallets, w)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	order := map[string]int{}
	for i, code := range domain.WalletCodes {
		order[code] = i
	}
	sort.Slice(wallets, func(i, j int) bool { return order[wallets[i].Code] < order[wallets[j].Code] })
	available, equity, positions := decimal.Zero, decimal.Zero, decimal.Zero
	for _, w := range wallets {
		v := valueWallet(w, names, quotes, out.AsOf)
		if v.Decisions, e = readJSON(ctx, tx, "SELECT jsonb_build_object('id',d.id,'state',CASE WHEN d.state<>'completed' AND j.state IN ('failed','expired') THEN j.state ELSE d.state END,'started_at',d.started_at,'finished_at',d.finished_at,'model',d.model,'output',d.output) FROM decision_runs d LEFT JOIN jobs j ON j.id=d.job_id WHERE d.wallet_id=$1 ORDER BY d.started_at DESC LIMIT 10", w.ID); e != nil {
			return out, e
		}
		if v.Fills, e = readJSON(ctx, tx, "SELECT body FROM fills WHERE wallet_id=$1 ORDER BY occurred_at DESC LIMIT 100", w.ID); e != nil {
			return out, e
		}
		out.Wallets = append(out.Wallets, v)
		available = available.Add(val(v.Available))
		if !v.Complete {
			out.Complete = false
		} else {
			equity = equity.Add(val(*v.Equity))
			positions = positions.Add(val(*v.PositionValue))
		}
	}
	out.Available = available.String()
	if out.Complete {
		out.Equity = ptrMoney(equity)
		out.PositionValue = ptrMoney(positions)
	}
	if e = tx.QueryRow(ctx, "SELECT paused,execution_enabled,live_acceptance FROM runtime_control WHERE id").Scan(&out.Paused, &out.Execution, &out.LiveAcceptance); e != nil {
		return out, e
	}
	if e = tx.QueryRow(ctx, "SELECT count(*) FROM receipts WHERE delivered_at IS NULL").Scan(&out.PendingReceipts); e != nil {
		return out, e
	}
	if out.Jobs, e = readJSON(ctx, tx, "SELECT jsonb_build_object('kind',kind,'state',state,'available_at',available_at,'deadline',deadline,'last_error',last_error) FROM jobs ORDER BY available_at DESC LIMIT 40"); e != nil {
		return out, e
	}
	if out.Events, e = readJSON(ctx, tx, "SELECT to_jsonb(t) FROM system_events t ORDER BY id DESC LIMIT 30"); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func readJSON(ctx context.Context, tx pgx.Tx, sql string, args ...any) ([]json.RawMessage, error) {
	out := []json.RawMessage{}
	rows, e := tx.Query(ctx, sql, args...)
	if e != nil {
		return out, e
	}
	defer rows.Close()
	for rows.Next() {
		var b []byte
		if e = rows.Scan(&b); e != nil {
			return out, e
		}
		out = append(out, json.RawMessage(b))
	}
	return out, rows.Err()
}
