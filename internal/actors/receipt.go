package actors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/contracts"
	"github.com/caglarkc/trading/internal/bridge"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/modelrunner"
	"github.com/google/uuid"
	"time"
)

func (s *Service) applyAndAcknowledge(ctx context.Context, j jobs.Job, did uuid.UUID, wallet string, scope bridge.Scope, result modelrunner.Result) (any, error) {
	if e := contracts.Validate(contracts.DecisionSchema(), result.Output); e != nil {
		return nil, e
	}
	var decision domain.Decision
	dec := json.NewDecoder(bytes.NewReader(result.Output))
	dec.DisallowUnknownFields()
	if e := dec.Decode(&decision); e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	for _, a := range decision.Actions {
		if seen[a.ClientID] {
			return nil, fmt.Errorf("duplicate_client_action")
		}
		for _, dep := range a.Depends {
			if !seen[dep] {
				return nil, fmt.Errorf("invalid_action_dependency")
			}
		}
		evidence := map[string]bool{}
		for _, id := range a.EvidenceIDs {
			if _, err := uuid.Parse(id); err != nil || evidence[id] {
				return nil, fmt.Errorf("invalid_or_duplicate_evidence_id")
			}
			evidence[id] = true
		}
		seen[a.ClientID] = true
	}
	tx, e := s.Store.Pool.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if e = jobs.Guard(ctx, tx, j); e != nil {
		return nil, e
	}
	_, e = tx.Exec(ctx, "UPDATE decision_runs SET output=$3,output_bytes=$9,output_hash=$4,runtime_events=$5,duration_ms=$6,usage=$7,state='ready',session_id=$8 WHERE wallet_id=$1 AND id=$2", wallet, did, result.Output, digest(result.Output), bytesOf(result.Events), result.DurationMS, result.Usage, result.SessionID, []byte(result.Output))
	if e != nil {
		return nil, e
	}
	if e = tx.Commit(ctx); e != nil {
		return nil, e
	}
	receipts := []json.RawMessage{}
	for _, a := range decision.Actions {
		receipt, e := bridge.Submit(ctx, scope, a)
		if e != nil {
			return nil, e
		}
		receipts = append(receipts, receipt)
	}
	if e = s.Store.SaveMemory(ctx, wallet, did.String(), decision.Memory, j); e != nil {
		return nil, e
	}
	rows, e := s.Store.Pool.Query(ctx, "SELECT id,body FROM receipts WHERE wallet_id=$1 AND delivered_at IS NULL ORDER BY id LIMIT 200", wallet)
	if e != nil {
		return nil, e
	}
	type delivery struct {
		ID   int64           `json:"receipt_id"`
		Body json.RawMessage `json:"receipt"`
	}
	pending := []delivery{}
	allowed := map[int64]bool{}
	for rows.Next() {
		var d delivery
		if e = rows.Scan(&d.ID, &d.Body); e != nil {
			rows.Close()
			return nil, e
		}
		allowed[d.ID] = true
		pending = append(pending, d)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	ackID := uuid.NewSHA1(did, []byte("receipt-v1"))
	prompt := `These are authoritative account receipts for your already delivered action batch. An accepted or waiting order is not filled. A rejection or partial fill is not success. Read your current account with the scoped tool if needed. Do not issue a new action batch or change character. Acknowledge only provided receipt IDs; update your own method/monitoring notes using actual outcomes, distinguishing opinion from fact. Return only the receipt schema. ` + string(bytesOf(pending))
	ack, e := s.Runner.Run(ctx, modelrunner.Request{Wallet: wallet, SessionID: result.SessionID, Prompt: prompt, Schema: contracts.ReceiptSchema(), MCP: true, Web: false, RunID: ackID.String(), Deadline: time.Now().Add(6 * time.Minute)})
	if e != nil {
		return nil, e
	}
	var response struct {
		IDs    []int64 `json:"acknowledged_receipt_ids"`
		Memory string  `json:"memory_update"`
	}
	if e = json.Unmarshal(ack.Output, &response); e != nil {
		return nil, e
	}
	for _, id := range response.IDs {
		if !allowed[id] {
			return nil, fmt.Errorf("receipt_scope_mismatch")
		}
		delete(allowed, id)
	}
	if len(response.IDs) != len(pending) {
		return nil, fmt.Errorf("receipt_ack_incomplete")
	}
	if e = s.Store.SaveMemory(ctx, wallet, did.String(), response.Memory, j); e != nil {
		return nil, e
	}
	tx, e = s.Store.Pool.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if e = jobs.Guard(ctx, tx, j); e != nil {
		return nil, e
	}
	_, e = tx.Exec(ctx, "UPDATE receipts SET delivered_at=now() WHERE wallet_id=$1 AND id=ANY($2) AND delivered_at IS NULL", wallet, response.IDs)
	if e != nil {
		return nil, e
	}
	_, e = tx.Exec(ctx, "UPDATE decision_runs SET state='completed',finished_at=now() WHERE wallet_id=$1 AND id=$2", wallet, did)
	if e != nil {
		return nil, e
	}
	if e = tx.Commit(ctx); e != nil {
		return nil, e
	}
	return map[string]any{"decision_id": did, "session_id": result.SessionID, "actions": len(decision.Actions), "receipts_acknowledged": len(response.IDs), "output_hash": digest(result.Output), "model": result.Model, "reasoning": result.Reasoning}, nil
}
