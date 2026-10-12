package store

import (
	"context"
	"fmt"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/google/uuid"
	"unicode/utf8"
)

func (s *Store) SaveMemory(ctx context.Context, wallet, decision, summary string, j jobs.Job) error {
	if utf8.RuneCountInString(summary) > 12000 {
		return fmt.Errorf("memory_too_large")
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	wid, e := uuid.Parse(wallet)
	if e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "SELECT id FROM wallets WHERE id=$1 FOR UPDATE", wid); e != nil {
		return e
	}
	if e = jobs.Guard(ctx, tx, j); e != nil {
		return e
	}
	var exists bool
	if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM memory_versions WHERE wallet_id=$1 AND decision_id=$2 AND summary=$3)", wid, decision, summary).Scan(&exists); e != nil {
		return e
	}
	if exists {
		return tx.Commit(ctx)
	}
	var version int64
	e = tx.QueryRow(ctx, "SELECT coalesce(max(version),0)+1 FROM memory_versions WHERE wallet_id=$1", wid).Scan(&version)
	if e != nil {
		return e
	}
	var previous *int64
	if version > 1 {
		v := version - 1
		previous = &v
	}
	_, e = tx.Exec(ctx, "INSERT INTO memory_versions(wallet_id,version,decision_id,previous_version,summary) VALUES($1,$2,$3,$4,$5)", wid, version, decision, previous, summary)
	if e != nil {
		return e
	}
	return tx.Commit(ctx)
}
