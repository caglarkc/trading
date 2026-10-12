package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"time"
)

// Reconstruct ex-date ownership from chronological executions and splits. Current
// holdings cannot determine entitlement after an ex-date sale.
func historicalShares(ctx context.Context, tx pgx.Tx, wallet, asset string, cutoff time.Time) (int64, error) {
	rows, e := tx.Query(ctx, `SELECT at,kind,amount FROM (
 SELECT f.occurred_at AS at,CASE WHEN o.body->>'side'='BUY' THEN 'BUY' ELSE 'SELL' END AS kind, f.body->>'quantity' AS amount FROM fills f JOIN orders o ON o.wallet_id=f.wallet_id AND o.id=f.order_id WHERE f.wallet_id=$1 AND o.asset_id=$2 AND f.occurred_at<$3
 UNION ALL SELECT effective_at,'SPLIT',body->>'ratio' FROM corporate_events WHERE asset_id=$2 AND event_type='SPLIT' AND status='verified' AND effective_at<$3
 ) t ORDER BY at,CASE WHEN kind='SPLIT' THEN 0 ELSE 1 END`, wallet, asset, cutoff)
	if e != nil {
		return 0, e
	}
	defer rows.Close()
	shares := decimal.Zero
	for rows.Next() {
		var at time.Time
		var kind, amount string
		if e = rows.Scan(&at, &kind, &amount); e != nil {
			return 0, e
		}
		d, e := decimal.NewFromString(amount)
		if e != nil {
			return 0, e
		}
		switch kind {
		case "BUY":
			shares = shares.Add(d)
		case "SELL":
			shares = shares.Sub(d)
		case "SPLIT":
			shares = shares.Mul(d)
		}
	}
	if e = rows.Err(); e != nil {
		return 0, e
	}
	if shares.IsNegative() || !shares.Equal(shares.Truncate(0)) {
		return 0, fmt.Errorf("historical_entitlement_unresolved")
	}
	return shares.IntPart(), nil
}

// Keeps headers and typed event payload consistent before trusted application.
func eventIdentity(body []byte, id, asset string) error {
	var v struct{ ID, AssetID string }
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(body, &fields); e != nil {
		return e
	}
	json.Unmarshal(fields["id"], &v.ID)
	json.Unmarshal(fields["asset_id"], &v.AssetID)
	if v.ID != id || v.AssetID != asset {
		return fmt.Errorf("corporate_event_identity_mismatch")
	}
	return nil
}
