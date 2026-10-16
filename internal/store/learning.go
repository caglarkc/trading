package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StrategyState struct {
	Version int64           `json:"version"`
	Through int64           `json:"reviewed_through_seq"`
	Report  json.RawMessage `json:"report"`
}
type LearningFill struct {
	ID    string          `json:"fill_id"`
	Seq   int64           `json:"seq"`
	Facts json.RawMessage `json:"facts"`
}

func (s *Store) LatestStrategy(ctx context.Context, wallet string) (StrategyState, error) {
	state := StrategyState{Report: json.RawMessage(`{}`)}
	err := s.Pool.QueryRow(ctx, `SELECT version,reviewed_through_seq,report FROM strategy_versions WHERE wallet_id=$1 ORDER BY version DESC LIMIT 1`, wallet).Scan(&state.Version, &state.Through, &state.Report)
	if err == pgx.ErrNoRows {
		err = nil
	}
	return state, err
}

// Realized P/L and disposed cost come from the immutable accounting journal,
// not a model's arithmetic or a reconstruction that could mishandle splits.
func (s *Store) LearningFills(ctx context.Context, wallet string, after int64) ([]LearningFill, error) {
	rows, err := s.Pool.Query(ctx, `SELECT f.id::text,f.seq,jsonb_build_object(
 'symbol',a.symbol,'fill',f.body,'exit_order',o.body,
 'net_realized_pnl_usd',(SELECT (-sum(l.amount))::text FROM journal_transactions t JOIN journal_lines l ON l.wallet_id=t.wallet_id AND l.transaction_id=t.id WHERE t.wallet_id=f.wallet_id AND t.event_type='execution' AND t.external_event_id=f.id::text AND l.account_code='realized_income'),
 'disposed_cost_basis_usd',(SELECT (-sum(l.amount))::text FROM journal_transactions t JOIN journal_lines l ON l.wallet_id=t.wallet_id AND l.transaction_id=t.id WHERE t.wallet_id=f.wallet_id AND t.event_type='execution' AND t.external_event_id=f.id::text AND l.account_code='position_cost'),
 'exit_decision',(SELECT jsonb_build_object('action',(SELECT act FROM jsonb_array_elements(d.output->'actions') act WHERE act->>'client_action_id'=da.client_action_id LIMIT 1),'strategy_application',d.output->'learning_application','strategy_version',d.input->'learning'->'version') FROM decision_actions da JOIN decision_runs d ON d.wallet_id=da.wallet_id AND d.id=da.decision_id WHERE da.wallet_id=f.wallet_id AND da.receipt->'order'->>'id'=o.id::text ORDER BY da.created_at LIMIT 1),
 'entry_history_not_fifo_attribution',coalesce((SELECT jsonb_agg(b) FROM (SELECT bf.body AS fill,bo.body->>'thesis' AS thesis,
 (SELECT (SELECT act FROM jsonb_array_elements(d.output->'actions') act WHERE act->>'client_action_id'=da.client_action_id LIMIT 1) FROM decision_actions da JOIN decision_runs d ON d.wallet_id=da.wallet_id AND d.id=da.decision_id WHERE da.wallet_id=bf.wallet_id AND da.receipt->'order'->>'id'=bo.id::text ORDER BY da.created_at LIMIT 1) AS original_plan
 FROM fills bf JOIN orders bo ON bo.wallet_id=bf.wallet_id AND bo.id=bf.order_id WHERE bf.wallet_id=f.wallet_id AND bo.asset_id=o.asset_id AND bo.body->>'side'='BUY' AND bf.seq<f.seq ORDER BY bf.seq DESC LIMIT 10) b),'[]'::jsonb)
 ) FROM fills f JOIN orders o ON o.wallet_id=f.wallet_id AND o.id=f.order_id JOIN assets a ON a.id=o.asset_id
 WHERE f.wallet_id=$1 AND f.seq>$2 AND o.body->>'side'='SELL' ORDER BY f.seq LIMIT 20`, wallet, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LearningFill{}
	for rows.Next() {
		var f LearningFill
		if err = rows.Scan(&f.ID, &f.Seq, &f.Facts); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) SaveStrategy(ctx context.Context, wallet string, previous StrategyState, through int64, input, report json.RawMessage, session string, usage json.RawMessage, j jobs.Job) error {
	if j.Wallet == nil || j.Wallet.String() != wallet {
		return fmt.Errorf("strategy_wallet_scope_mismatch")
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = jobs.Guard(ctx, tx, j); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SELECT id FROM wallets WHERE id=$1 FOR UPDATE", wallet); err != nil {
		return err
	}
	if len(usage) == 0 {
		usage = json.RawMessage(`{}`)
	}
	tag, err := tx.Exec(ctx, `INSERT INTO strategy_versions(wallet_id,version,job_id,reviewed_through_seq,input,report,session_id,usage)
 SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE NOT EXISTS(SELECT 1 FROM strategy_versions WHERE wallet_id=$1 AND version>$9)`, uuid.MustParse(wallet), previous.Version+1, j.ID, through, input, report, session, usage, previous.Version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("strategy_revision_conflict")
	}
	return tx.Commit(ctx)
}
