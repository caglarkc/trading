package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Job struct {
	ID        uuid.UUID       `json:"id"`
	Kind      string          `json:"kind"`
	Key       string          `json:"key"`
	Wallet    *uuid.UUID      `json:"wallet_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Available time.Time       `json:"available_at"`
	Deadline  time.Time       `json:"deadline"`
	Owner     string          `json:"lease_owner"`
	Fence     int64           `json:"fencing_token"`
	Attempts  int             `json:"attempts"`
}
type Queue struct{ Pool *pgxpool.Pool }

func (q Queue) Enqueue(ctx context.Context, j Job, dependency *uuid.UUID, priority int) (uuid.UUID, error) {
	if j.ID == uuid.Nil {
		j.ID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:job:"+j.Kind+":"+j.Key))
	}
	if len(j.Payload) == 0 {
		j.Payload = []byte(`{}`)
	}
	var id uuid.UUID
	e := q.Pool.QueryRow(ctx, "INSERT INTO jobs(id,kind,job_key,wallet_id,payload,available_at,deadline,depends_on,priority) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(kind,job_key) DO UPDATE SET job_key=jobs.job_key RETURNING id", j.ID, j.Kind, j.Key, j.Wallet, j.Payload, j.Available, j.Deadline, dependency, priority).Scan(&id)
	return id, e
}
func (q Queue) Acquire(ctx context.Context, owner string, kinds []string) (*Job, error) {
	tx, e := q.Pool.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	_, e = tx.Exec(ctx, "UPDATE jobs SET state='expired',last_error='job_deadline',finished_at=now(),fencing_token=fencing_token+1 WHERE deadline<=now() AND state IN ('pending','running') AND (lease_until IS NULL OR lease_until<=now())")
	if e != nil {
		return nil, e
	}
	var j Job
	e = tx.QueryRow(ctx, `SELECT j.id,j.kind,j.job_key,j.wallet_id,j.payload,j.available_at,j.deadline,j.fencing_token,j.attempts FROM jobs j WHERE j.kind=ANY($1) AND j.available_at<=now() AND j.deadline>now() AND j.attempts<3 AND (j.state='pending' OR j.state='running' AND j.lease_until<=now()) AND (j.depends_on IS NULL OR EXISTS(SELECT 1 FROM jobs d WHERE d.id=j.depends_on AND d.state IN ('completed','failed','expired'))) AND (j.wallet_id IS NULL OR NOT EXISTS(SELECT 1 FROM jobs active WHERE active.wallet_id=j.wallet_id AND active.state='running' AND active.id<>j.id)) AND (j.kind<>'price_refresh' OR NOT EXISTS(SELECT 1 FROM jobs active WHERE active.kind='price_refresh' AND active.state='running' AND active.id<>j.id AND active.payload->>'asset_id'=j.payload->>'asset_id')) AND NOT EXISTS(SELECT 1 FROM runtime_control WHERE paused) ORDER BY j.priority,j.available_at,j.id FOR UPDATE OF j SKIP LOCKED LIMIT 1`, kinds).Scan(&j.ID, &j.Kind, &j.Key, &j.Wallet, &j.Payload, &j.Available, &j.Deadline, &j.Fence, &j.Attempts)
	if e == pgx.ErrNoRows {
		return nil, tx.Commit(ctx)
	}
	if e != nil {
		return nil, e
	}
	j.Owner = owner
	j.Fence++
	j.Attempts++
	_, e = tx.Exec(ctx, "UPDATE jobs SET state='running',lease_owner=$2,lease_until=least(now()+interval '60 seconds',deadline),fencing_token=$3,attempts=$4 WHERE id=$1", j.ID, owner, j.Fence, j.Attempts)
	if e != nil {
		var pe *pgconn.PgError
		if errors.As(e, &pe) && pe.Code == "23505" {
			return nil, nil
		}
		return nil, e
	}
	if e = tx.Commit(ctx); e != nil {
		return nil, e
	}
	return &j, nil
}
func (q Queue) Heartbeat(ctx context.Context, j Job) error {
	tag, e := q.Pool.Exec(ctx, "UPDATE jobs SET lease_until=least(now()+interval '60 seconds',deadline) WHERE id=$1 AND lease_owner=$2 AND fencing_token=$3 AND state='running' AND lease_until>now() AND deadline>now()", j.ID, j.Owner, j.Fence)
	if e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("lease_lost")
	}
	return nil
}
func (q Queue) Finish(ctx context.Context, j Job, result any, workErr error) error {
	state := "completed"
	var errText *string
	if workErr != nil {
		state = "failed"
		s := workErr.Error()
		if len(s) > 2000 {
			s = s[:2000]
		}
		errText = &s
	}
	b, e := json.Marshal(result)
	if e != nil {
		return e
	}
	tag, e := q.Pool.Exec(ctx, "UPDATE jobs SET state=$4,result=$5,last_error=$6,finished_at=now(),lease_until=NULL WHERE id=$1 AND lease_owner=$2 AND fencing_token=$3 AND state='running' AND lease_until>now() AND deadline>now()", j.ID, j.Owner, j.Fence, state, b, errText)
	if e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("lease_lost")
	}
	return nil
}

// Heartbeat failure cancels the work, including native CLI process and container.
func (q Queue) Run(ctx context.Context, j Job, fn func(context.Context) (any, error)) error {
	work, cancel := context.WithDeadline(ctx, j.Deadline)
	defer cancel()
	done := make(chan struct{})
	leaseErr := make(chan error, 1)
	go func() {
		timer := time.NewTicker(15 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-done:
				return
			case <-work.Done():
				return
			case <-timer.C:
				c, cancel := context.WithTimeout(work, 5*time.Second)
				e := q.Heartbeat(c, j)
				cancel()
				if e != nil {
					leaseErr <- e
					return
				}
			}
		}
	}()
	// Separate heartbeat loop below watches the error and immediately cancels fn.
	watchDone := make(chan struct{})
	go func() {
		select {
		case <-watchDone:
		case <-leaseErr:
			cancel()
		}
	}()
	result, e := fn(work)
	close(done)
	close(watchDone)
	if work.Err() != nil {
		return work.Err()
	}
	if finishErr := q.Finish(ctx, j, result, e); finishErr != nil {
		return finishErr
	}
	return e
}

// Guard holds a shared job-row lock inside the result transaction. Acquisition
// and heartbeat changes cannot race past a successful fencing check.
func Guard(ctx context.Context, tx pgx.Tx, j Job) error {
	var valid bool
	e := tx.QueryRow(ctx, "SELECT state='running' AND lease_owner=$2 AND fencing_token=$3 AND lease_until>clock_timestamp() AND deadline>clock_timestamp() FROM jobs WHERE id=$1 FOR SHARE", j.ID, j.Owner, j.Fence).Scan(&valid)
	if e != nil {
		return e
	}
	if !valid {
		return fmt.Errorf("lease_lost")
	}
	return nil
}
