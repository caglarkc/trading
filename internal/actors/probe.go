package actors

import (
	"context"
	"fmt"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/modelrunner"
	"github.com/google/uuid"
	"sync"
	"time"
)

func (s *Service) Probe(ctx context.Context, count int) error {
	unlock, e := modelrunner.CoordinatorLock(s.Root)
	if e != nil {
		return e
	}
	defer unlock()
	if e = s.Runner.Recover(ctx); e != nil {
		return e
	}
	if count != 2 && count != 8 {
		return fmt.Errorf("probe_count_invalid")
	}
	codes := domain.WalletCodes
	if count == 2 {
		codes = []string{"QQQ-FAST", "QQQ-PATIENT"}
	}
	kind := "wallet_probe-" + uuid.NewString()
	q := jobs.Queue{Pool: s.Store.Pool}
	for _, code := range codes {
		wid := domain.WalletID(code)
		if count == 8 {
			var done bool
			if e = s.Store.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM decision_runs WHERE wallet_id=$1 AND state='completed')", wid).Scan(&done); e != nil {
				return e
			}
			if done {
				continue
			}
		}
		_, e := q.Enqueue(ctx, jobs.Job{Kind: kind, Key: uuid.NewString(), Wallet: &wid, Available: time.Now().Add(-time.Second), Deadline: time.Now().Add(40 * time.Minute)}, nil, 1)
		if e != nil {
			return e
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error, len(codes))
	for slot := range 2 {
		wg.Add(1)
		go func(slot int) {
			defer wg.Done()
			for {
				j, e := q.Acquire(ctx, fmt.Sprintf("wallet-probe-%d", slot), []string{kind})
				if e != nil {
					errs <- e
					return
				}
				if j == nil {
					return
				}
				if e = q.Run(ctx, *j, func(c context.Context) (any, error) { return s.Run(c, *j) }); e != nil {
					errs <- e
					return
				}
			}
		}(slot)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
