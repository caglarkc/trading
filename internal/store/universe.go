package store

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

// Universe activation and cancellation commit together. Removed holdings remain
// sellable; only new buys and unfilled buy quantities are excluded.
func CancelRemovedBuys(ctx context.Context, tx pgx.Tx, version uuid.UUID) error {
	rows, e := tx.Query(ctx, "SELECT id FROM wallets WHERE category='US' ORDER BY id FOR UPDATE")
	if e != nil {
		return e
	}
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	q := dbgen.New(tx)
	for _, id := range ids {
		row, e := q.GetWallet(ctx, id)
		if e != nil {
			return e
		}
		var a domain.Account
		if e = json.Unmarshal(row.State, &a); e != nil {
			return e
		}
		before := clone(a)
		changed := false
		for _, o := range a.Orders {
			if o.Side != "BUY" || !domain.Active(o) {
				continue
			}
			var allowed bool
			if e = tx.QueryRow(ctx, "SELECT fixed OR EXISTS(SELECT 1 FROM dynamic_members WHERE version_id=$2 AND asset_id=assets.id) FROM assets WHERE id=$1", o.AssetID, version).Scan(&allowed); e != nil {
				return e
			}
			if allowed {
				continue
			}
			if e = a.Cancel(o.ID); e != nil {
				return e
			}
			changed = true
			receipt := domain.Receipt{Status: "cancelled", Error: "dynamic_universe_removed", Order: o, AsOf: time.Now().UTC()}
			if _, e = tx.Exec(ctx, "INSERT INTO receipts(wallet_id,body) VALUES($1,$2)", id, marshal(receipt)); e != nil {
				return e
			}
		}
		if changed {
			if e = persist(ctx, tx, q, id, before, a, "universe_activation", version.String(), time.Now().UTC()); e != nil {
				return e
			}
		}
	}
	return nil
}
