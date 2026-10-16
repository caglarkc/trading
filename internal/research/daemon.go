package research

import (
	"context"
	"fmt"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/modelrunner"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func (s *Service) Loop(ctx context.Context, root string, probe bool) error {
	unlock, e := modelrunner.CoordinatorLock(root)
	if e != nil {
		return e
	}
	defer unlock()
	if e = s.Runner.Recover(ctx); e != nil {
		return e
	}
	q := jobs.Queue{Pool: s.Pool}
	owner := "model-" + uuid.NewString()
	if probe {
		j := jobs.Job{Kind: "research_probe", Key: uuid.NewString(), Available: time.Now().Add(-time.Second), Deadline: time.Now().Add(5 * time.Minute)}
		if _, e = q.Enqueue(ctx, j, nil, 3); e != nil {
			return e
		}
		claimed, e := q.Acquire(ctx, owner, []string{"research_probe"})
		if e != nil || claimed == nil {
			return fmt.Errorf("research_probe_not_acquired: %v", e)
		}
		return q.Run(ctx, *claimed, func(c context.Context) (any, error) { return s.Run(c, *claimed) })
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	for slot := range 2 {
		wg.Add(1)
		go func(slot int) {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				kinds := []string{"research", "dynamic_universe", "wallet_decision", "wallet_screen", "wallet_deep", "wallet_allocation", "wallet_learning"}
				if _, err := os.Stat(filepath.Join(filepath.Dir(root), "reviews", "research-only.lock")); err == nil {
					kinds = []string{"wallet_screen", "wallet_deep", "wallet_allocation"}
				}
				j, e := q.Acquire(ctx, fmt.Sprintf("%s-%d", owner, slot), kinds)
				if e != nil {
					errCh <- e
					return
				}
				if j == nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(5 * time.Second):
					}
					continue
				}
				if e = q.Run(ctx, *j, func(c context.Context) (any, error) {
					if j.Kind == "dynamic_universe" {
						return s.Dynamic(c, *j)
					}
					if j.Kind == "wallet_decision" || j.Kind == "wallet_screen" || j.Kind == "wallet_deep" || j.Kind == "wallet_allocation" || j.Kind == "wallet_learning" {
						if s.WalletHandler == nil {
							return nil, fmt.Errorf("wallet_handler_unavailable")
						}
						return s.WalletHandler(c, *j)
					}
					return s.Run(c, *j)
				}); e != nil && ctx.Err() == nil {
					b := []byte(fmt.Sprintf(`{"kind":"research","error":"job_failed","job_id":"%s"}`, j.ID))
					s.Pool.Exec(ctx, "INSERT INTO system_events(kind,body) VALUES('model_job_error',$1)", b)
				}
			}
		}(slot)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case e := <-errCh:
		return e
	case <-ctx.Done():
		<-done
		return nil
	}
}
