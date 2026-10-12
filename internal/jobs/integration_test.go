package jobs_test

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/database"
	. "github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/store"
	"github.com/google/uuid"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDurableLeaseAndFencing(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	if e := database.Migrate(ctx, os.Getenv("TEST_OWNER_URL")); e != nil {
		t.Fatal(e)
	}
	s, e := store.Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	q := Queue{s.Pool}
	kind := "fixture-" + uuid.NewString()
	j := Job{Kind: kind, Key: uuid.NewString(), Available: time.Now().Add(-time.Second), Deadline: time.Now().Add(5 * time.Minute)}
	id, e := q.Enqueue(ctx, j, nil, 0)
	if e != nil {
		t.Fatal(e)
	}
	again, e := q.Enqueue(ctx, j, nil, 0)
	if e != nil || again != id {
		t.Fatal("duplicate enqueue", e)
	}
	var wg sync.WaitGroup
	results := make([]*Job, 2)
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) { defer wg.Done(); results[i], errs[i] = q.Acquire(ctx, uuid.NewString(), []string{kind}) }(i)
	}
	wg.Wait()
	var first *Job
	for i := range 2 {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if results[i] != nil {
			if first != nil {
				t.Fatal("duplicate lease")
			}
			first = results[i]
		}
	}
	if first == nil {
		t.Fatal("no job")
	}
	if _, e = s.Pool.Exec(ctx, "UPDATE jobs SET lease_until=now()-interval '1 second' WHERE id=$1", id); e != nil {
		t.Fatal(e)
	}
	second, e := q.Acquire(ctx, "restart-worker", []string{kind})
	if e != nil || second == nil || second.Fence <= first.Fence {
		t.Fatal(second, e)
	}
	if e = q.Heartbeat(ctx, *first); e == nil {
		t.Fatal("stale heartbeat accepted")
	}
	if e = q.Finish(ctx, *first, map[string]bool{"stale": true}, nil); e == nil {
		t.Fatal("stale completion accepted")
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if e = Guard(ctx, tx, *first); e == nil {
		t.Fatal("stale transaction accepted")
	}
	tx.Rollback(ctx)
	if e = q.Finish(ctx, *second, map[string]bool{"recovered": true}, nil); e != nil {
		t.Fatal(e)
	}
	if next, e := q.Acquire(ctx, "after-restart", []string{kind}); e != nil || next != nil {
		t.Fatal("completed job replay", e)
	}
	expired := Job{Kind: kind, Key: uuid.NewString(), Available: time.Now().Add(-2 * time.Minute), Deadline: time.Now().Add(-time.Minute)}
	eid, e := q.Enqueue(ctx, expired, nil, 1)
	if e != nil {
		t.Fatal(e)
	}
	q.Acquire(ctx, "expiry", []string{kind})
	var state string
	if e = s.Pool.QueryRow(ctx, "SELECT state FROM jobs WHERE id=$1", eid).Scan(&state); e != nil || state != "expired" {
		t.Fatal(state, e)
	}
	t.Log("enqueue replay, SKIP LOCKED, restart lease reclamation, stale heartbeat/result/transaction rejection and deadline expiration passed")
}
func TestSessionPlan(t *testing.T) {
	if os.Getenv("TEST_OWNER_URL") == "" {
		t.Skip("isolated PostgreSQL required")
	}
	ctx := context.Background()
	s, e := store.Open(ctx, os.Getenv("TEST_WORKER_URL"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	cal, e := calendar.Load("../../config/calendar.json")
	if e != nil {
		t.Fatal(e)
	}
	p := Planner{Queue{s.Pool}, cal}
	now := time.Date(2026, 11, 27, 14, 0, 0, 0, time.UTC)
	if e = p.Plan(ctx, now); e != nil {
		t.Fatal(e)
	}
	if e = p.Plan(ctx, now); e != nil {
		t.Fatal(e)
	}
	rows, e := s.Pool.Query(ctx, "SELECT available_at,deadline,payload FROM jobs WHERE kind='wallet_decision' AND job_key LIKE '2026-11-27%' ORDER BY available_at")
	if e != nil {
		t.Fatal(e)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var at, deadline time.Time
		var body json.RawMessage
		if e = rows.Scan(&at, &deadline, &body); e != nil {
			t.Fatal(e)
		}
		if at.Format("15:04") != "15:00" && at.Format("15:04") != "17:00" {
			t.Fatal("wrong session", at)
		}
		if deadline.Sub(at) != 45*time.Minute {
			t.Fatal("wrong deadline", deadline)
		}
		count++
	}
	if count != 16 {
		t.Fatal(count)
	}
	t.Log("two sessions per eight wallets, early close and idempotent planning passed")
}
