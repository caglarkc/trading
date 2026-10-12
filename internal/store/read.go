package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Rows selects only a fixed internal table expression. Caller text is never SQL.
func (s *Store) Rows(ctx context.Context, kind, wallet string, cursor int64, limit int) ([]json.RawMessage, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	clauses := map[string]string{
		"orders":   "SELECT body || jsonb_build_object('cursor',seq) FROM orders WHERE ($1='' OR wallet_id=nullif($1,'')::uuid) AND ($2=0 OR seq<$2) ORDER BY seq DESC LIMIT $3",
		"fills":    "SELECT body || jsonb_build_object('cursor',seq) FROM fills WHERE ($1='' OR wallet_id=nullif($1,'')::uuid) AND ($2=0 OR seq<$2) ORDER BY seq DESC LIMIT $3",
		"receipts": "SELECT to_jsonb(t) FROM receipts t WHERE wallet_id=nullif($1,'')::uuid AND id>$2 ORDER BY id LIMIT $3",
		"evidence": "SELECT to_jsonb(t) FROM evidence t WHERE $1='' AND extract(epoch from observed_at)::bigint >= $2 ORDER BY observed_at DESC,id LIMIT $3",
		"events":   "SELECT to_jsonb(t) FROM system_events t WHERE $1='' AND id>$2 ORDER BY id DESC LIMIT $3",
	}
	sql, ok := clauses[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported_read")
	}
	out := []json.RawMessage{}
	query := func(db interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	}) error {
		rows, e := db.Query(ctx, sql, wallet, cursor, limit)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var b []byte
			if e = rows.Scan(&b); e != nil {
				return e
			}
			out = append(out, json.RawMessage(b))
		}
		return rows.Err()
	}
	if wallet != "" {
		id, e := uuid.Parse(wallet)
		if e != nil {
			return nil, e
		}
		tx, e := s.Pool.Begin(ctx)
		if e != nil {
			return nil, e
		}
		defer tx.Rollback(ctx)
		if _, e = tx.Exec(ctx, "SELECT set_config('app.wallet_id',$1,true)", id.String()); e != nil {
			return nil, e
		}
		e = query(tx)
		if e == nil {
			e = tx.Commit(ctx)
		}
		return out, e
	}
	return out, query(s.Pool)
}
func (s *Store) LatestMemory(ctx context.Context, wallet string) (json.RawMessage, error) {
	var b []byte
	e := s.Scoped(ctx, wallet, func(tx pgx.Tx, _ *dbgen.Queries) error {
		e := tx.QueryRow(ctx, "SELECT to_jsonb(t) FROM memory_versions t WHERE wallet_id=$1 ORDER BY version DESC LIMIT 1", wallet).Scan(&b)
		if e == pgx.ErrNoRows {
			b = []byte(`{"version":0,"summary":""}`)
			return nil
		}
		return e
	})
	return json.RawMessage(b), e
}
