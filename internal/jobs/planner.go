package jobs

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/google/uuid"
	"time"
)

type Planner struct {
	Queue    Queue
	Calendar calendar.Calendar
}

func payload(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func (p Planner) Plan(ctx context.Context, now time.Time) error {
	bucket := now.UTC().Truncate(time.Minute)
	if _, e := p.Queue.Enqueue(ctx, Job{Kind: "settlement", Key: bucket.Format(time.RFC3339), Available: bucket, Deadline: bucket.Add(2 * time.Minute)}, nil, 2); e != nil {
		return e
	}
	rows, err := p.Queue.Pool.Query(ctx, "SELECT id,expected_at FROM calendar_events WHERE status='verified_time' AND expected_at BETWEEN $1 AND $2", now.Add(-20*time.Minute), now.Add(24*time.Hour))
	if err != nil {
		return err
	}
	type calendarEvent struct {
		id uuid.UUID
		at time.Time
	}
	events := []calendarEvent{}
	for rows.Next() {
		var e calendarEvent
		if err = rows.Scan(&e.id, &e.at); err != nil {
			rows.Close()
			return err
		}
		events = append(events, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, e := range events {
		if err = p.EventJobs(ctx, e.id, e.at); err != nil {
			return err
		}
	}
	rows, err = p.Queue.Pool.Query(ctx, "SELECT id,effective_at FROM corporate_events WHERE effective_at BETWEEN $1 AND $2", now.Add(-24*time.Hour), now.Add(24*time.Hour))
	if err != nil {
		return err
	}
	events = nil
	for rows.Next() {
		var e calendarEvent
		if err = rows.Scan(&e.id, &e.at); err != nil {
			rows.Close()
			return err
		}
		events = append(events, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, e := range events {
		if _, err = p.Queue.Enqueue(ctx, Job{Kind: "corporate_action", Key: e.id.String(), Available: e.at, Deadline: e.at.Add(24 * time.Hour), Payload: payload(map[string]string{"event_id": e.id.String()})}, nil, 0); err != nil {
			return err
		}
	}
	open, close, day, e := p.Calendar.Session(now)
	if e != nil || !day {
		return e
	}
	date := open.In(calendar.ET).Format("2006-01-02")
	// Leave the opening half hour for price discovery and a full hour before
	// close for the final research, decisions and order execution.
	for at := open.Add(30 * time.Minute); !at.After(close.Add(-time.Hour)); at = at.Add(time.Hour) {
		session := at.Format(time.RFC3339)
		deadline := at.Add(45 * time.Minute)
		if close.Before(deadline) {
			deadline = close
		}
		rid, e := p.Queue.Enqueue(ctx, Job{Kind: "research", Key: session, Available: at, Deadline: at.Add(5 * time.Minute), Payload: payload(map[string]any{"session": session})}, nil, 5)
		if e != nil {
			return e
		}
		for _, code := range domain.WalletCodes {
			id := domain.WalletID(code)
			if _, e = p.Queue.Enqueue(ctx, Job{Kind: "wallet_decision", Key: session + ":" + code, Wallet: &id, Available: at, Deadline: deadline, Payload: payload(map[string]any{"research_job_id": rid})}, &rid, 10); e != nil {
				return e
			}
		}
	}
	first := true
	for d := open.In(calendar.ET).AddDate(0, 0, -1); d.Weekday() != time.Sunday; d = d.AddDate(0, 0, -1) {
		_, _, trading, e := p.Calendar.Session(d)
		if e != nil {
			return e
		}
		if trading {
			first = false
			break
		}
	}
	if first {
		at := open.Add(-30 * time.Minute)
		year, week := at.In(calendar.ET).ISOWeek()
		key := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + ":" + stringWeek(week)
		if _, e = p.Queue.Enqueue(ctx, Job{Kind: "dynamic_universe", Key: key, Available: at, Deadline: open, Payload: payload(map[string]any{"week_key": key})}, nil, 4); e != nil {
			return e
		}
	}
	if _, e = p.Queue.Enqueue(ctx, Job{Kind: "close_accounting", Key: date, Available: close.Add(10 * time.Minute), Deadline: close.Add(40 * time.Minute)}, nil, 20); e != nil {
		return e
	}
	if !p.Calendar.Open(now) {
		return nil
	}
	rows, e = p.Queue.Pool.Query(ctx, `SELECT asset_id::text, bool_or(body->>'order_type'='MARKET' OR body->>'order_type'='STOP_MARKET' AND (body->>'triggered')::boolean) FROM orders WHERE state IN ('accepted','waiting','partially_filled') GROUP BY asset_id`)
	if e != nil {
		return e
	}
	type assetJob struct {
		id     string
		market bool
	}
	assets := []assetJob{}
	for rows.Next() {
		var a assetJob
		if e = rows.Scan(&a.id, &a.market); e != nil {
			rows.Close()
			return e
		}
		assets = append(assets, a)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, a := range assets {
		frequency := time.Minute
		priority := 1
		if a.market {
			frequency = 30 * time.Second
			priority = 0
		}
		slot := now.Truncate(frequency)
		if _, e = p.Queue.Enqueue(ctx, Job{Kind: "price_refresh", Key: a.id + ":" + slot.Format(time.RFC3339), Available: slot, Deadline: slot.Add(45 * time.Second), Payload: payload(map[string]any{"asset_id": a.id, "execute": true})}, nil, priority); e != nil {
			return e
		}
	}
	rows, e = p.Queue.Pool.Query(ctx, "SELECT DISTINCT asset_id::text FROM position_lots WHERE quantity>0")
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var asset string
		if e = rows.Scan(&asset); e != nil {
			return e
		}
		slot := now.Truncate(5 * time.Minute)
		if _, e = p.Queue.Enqueue(ctx, Job{Kind: "price_refresh", Key: asset + ":" + slot.Format(time.RFC3339), Available: slot, Deadline: slot.Add(5 * time.Minute), Payload: payload(map[string]any{"asset_id": asset, "execute": false})}, nil, 15); e != nil {
			return e
		}
	}
	return rows.Err()
}
func stringWeek(w int) string {
	const digits = "0123456789"
	return string([]byte{digits[w/10], digits[w%10]})
}
func (p Planner) EventJobs(ctx context.Context, id uuid.UUID, at time.Time) error {
	for _, delta := range []time.Duration{-30 * time.Minute, 5 * time.Minute, 15 * time.Minute} {
		t := at.Add(delta)
		if _, e := p.Queue.Enqueue(ctx, Job{Kind: "event_check", Key: id.String() + ":" + t.Format(time.RFC3339), Available: t, Deadline: t.Add(5 * time.Minute), Payload: payload(map[string]any{"event_id": id})}, nil, 8); e != nil {
			return e
		}
	}
	return nil
}
