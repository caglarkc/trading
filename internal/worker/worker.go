package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/caglarkc/trading/internal/store"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	Store         *store.Store
	Reader        *marketdata.Reader
	Calendar      calendar.Calendar
	Fees          domain.FeeProfile
	Execution     bool
	documentMu    sync.Mutex
	documents     map[string]CachedDocument
	documentLocks map[string]*sync.Mutex
}
type CachedDocument struct {
	ID       uuid.UUID           `json:"id"`
	Document marketdata.Document `json:"document"`
	Origin   string              `json:"origin_key"`
	Status   string              `json:"status"`
}

func (w *Worker) Loop(ctx context.Context) error {
	q := jobs.Queue{Pool: w.Store.Pool}
	planner := jobs.Planner{Queue: q, Calendar: w.Calendar}
	owner := "collector-" + uuid.NewString()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	nextPlan := time.Time{}
	for {
		if ctx.Err() != nil {
			return nil
		}
		if time.Now().After(nextPlan) {
			if e := planner.Plan(ctx, time.Now().UTC()); e != nil {
				w.event(ctx, "planner_error", map[string]string{"error": e.Error()})
			}
			nextPlan = time.Now().Add(15 * time.Second)
		}
		j, e := q.Acquire(ctx, owner, []string{"price_refresh", "settlement", "close_accounting", "corporate_action", "event_check"})
		if e != nil {
			return e
		}
		if j == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
			continue
		}
		e = q.Run(ctx, *j, func(c context.Context) (any, error) { return w.handle(c, *j) })
		if e != nil {
			w.event(ctx, "job_error", map[string]any{"job_id": j.ID, "kind": j.Kind, "error": e.Error()})
		}
	}
}
func (w *Worker) event(ctx context.Context, kind string, body any) {
	b, _ := json.Marshal(body)
	_, e := w.Store.Pool.Exec(ctx, "INSERT INTO system_events(kind,body) VALUES($1,$2)", kind, b)
	if e != nil {
		slog.Error("event_write_failed", "kind", kind)
	}
}
func (w *Worker) handle(ctx context.Context, j jobs.Job) (any, error) {
	switch j.Kind {
	case "price_refresh":
		var in struct {
			AssetID string `json:"asset_id"`
		}
		if e := json.Unmarshal(j.Payload, &in); e != nil {
			return nil, e
		}
		var a marketdata.Asset
		var aliases []byte
		e := w.Store.Pool.QueryRow(ctx, "SELECT id::text,symbol,name,category,exchange,aliases FROM assets WHERE id=$1", in.AssetID).Scan(&a.ID, &a.Symbol, &a.Name, &a.Category, &a.Exchange, &aliases)
		if e != nil {
			return nil, e
		}
		json.Unmarshal(aliases, &a.Aliases)
		c, cancel := context.WithTimeout(ctx, 45*time.Second)
		snapshot := w.Reader.Collect(c, a, w.Calendar.Open(time.Now()))
		cancel()
		b, _ := json.Marshal(snapshot)
		tx, e := w.Store.Pool.Begin(ctx)
		if e != nil {
			return nil, e
		}
		defer tx.Rollback(ctx)
		if e = jobs.Guard(ctx, tx, j); e != nil {
			return nil, e
		}
		for _, o := range snapshot.Observations {
			ob, _ := json.Marshal(o)
			if _, e = tx.Exec(ctx, "INSERT INTO observations(id,asset_id,source_id,event_time,observed_at,body) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO NOTHING", o.ID, a.ID, o.SourceID, o.EventTime, o.ObservedAt, ob); e != nil {
				return nil, e
			}
		}
		_, e = tx.Exec(ctx, "INSERT INTO market_snapshots(id,asset_id,captured_at,body) VALUES($1,$2,$3,$4)", uuid.New(), a.ID, time.Now().UTC(), b)
		if e != nil {
			return nil, e
		}
		if e = tx.Commit(ctx); e != nil {
			return nil, e
		}
		rows, e := w.Store.Pool.Query(ctx, "SELECT wallet_id::text,id::text FROM orders WHERE asset_id=$1 AND state IN ('accepted','waiting','partially_filled') ORDER BY wallet_id,id", a.ID)
		if e != nil {
			return nil, e
		}
		type candidate struct{ wallet, order string }
		list := []candidate{}
		for rows.Next() {
			var c candidate
			if e = rows.Scan(&c.wallet, &c.order); e != nil {
				rows.Close()
				return nil, e
			}
			list = append(list, c)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return nil, e
		}
		var accepted bool
		e = w.Store.Pool.QueryRow(ctx, "SELECT execution_enabled FROM runtime_control WHERE id").Scan(&accepted)
		if e != nil {
			return nil, e
		}
		for _, c := range list {
			if _, e = w.Store.Execute(ctx, c.wallet, c.order, snapshot, time.Now().UTC(), w.Calendar, w.Fees, w.Execution && accepted, j); e != nil {
				return nil, e
			}
		}
		w.event(ctx, "source_snapshot", map[string]any{"asset_id": a.ID, "symbol": a.Symbol, "price_verified": snapshot.Validation.Verified, "liquidity_verified": snapshot.LiquidityVerified, "validation_reason": snapshot.Validation.Reason, "observations": snapshot.Observations})
		return map[string]any{"asset_id": a.ID, "price_verified": snapshot.Validation.Verified, "liquidity_verified": snapshot.LiquidityVerified}, nil
	case "settlement", "close_accounting":
		wallets, e := w.Store.Wallets(ctx)
		if e != nil {
			return nil, e
		}
		for _, wallet := range wallets {
			if e = w.Store.Settle(ctx, wallet.ID.String(), time.Now().UTC(), j); e != nil {
				return nil, e
			}
			if j.Kind == "close_accounting" {
				if e = w.Store.Reconcile(ctx, wallet.ID.String()); e != nil {
					return nil, e
				}
			}
		}
		if j.Kind == "close_accounting" {
			w.event(ctx, "close_reconciliation", map[string]any{"wallets": len(wallets), "status": "passed"})
		}
		return map[string]any{"wallets": len(wallets)}, nil
	case "corporate_action":
		var in struct {
			EventID string `json:"event_id"`
		}
		if e := json.Unmarshal(j.Payload, &in); e != nil {
			return nil, e
		}
		var asset string
		if e := w.Store.Pool.QueryRow(ctx, "SELECT asset_id::text FROM corporate_events WHERE id=$1", in.EventID).Scan(&asset); e != nil {
			return nil, e
		}
		w.Store.Pool.Exec(ctx, "UPDATE assets SET execution_blocked=true WHERE id=$1", asset)
		wallets, e := w.Store.Wallets(ctx)
		if e != nil {
			return nil, e
		}
		for _, wallet := range wallets {
			if e = w.Store.Corporate(ctx, wallet.ID.String(), in.EventID, time.Now().UTC(), j); e != nil {
				return nil, e
			}
		}
		_, e = w.Store.Pool.Exec(ctx, "UPDATE assets SET execution_blocked=false WHERE id=$1", asset)
		return map[string]any{"event_id": in.EventID}, e
	case "event_check":
		var in struct {
			EventID string `json:"event_id"`
		}
		json.Unmarshal(j.Payload, &in)
		var source string
		if e := w.Store.Pool.QueryRow(ctx, "SELECT source_url FROM calendar_events WHERE id=$1", in.EventID).Scan(&source); e != nil {
			return nil, e
		}
		d, e := w.ReadDocument(ctx, source)
		if e != nil && j.Attempts == 1 && !strings.HasSuffix(j.Key, ":extra") {
			q := jobs.Queue{Pool: w.Store.Pool}
			at := time.Now().Add(30 * time.Minute)
			q.Enqueue(ctx, jobs.Job{Kind: "event_check", Key: j.Key + ":extra", Available: at, Deadline: at.Add(5 * time.Minute), Payload: j.Payload}, nil, 8)
		}
		return d, e
	}
	return nil, fmt.Errorf("unsupported_job_kind")
}
func (w *Worker) ReadDocument(ctx context.Context, target string) (CachedDocument, error) {
	w.documentMu.Lock()
	if w.documents == nil {
		w.documents = map[string]CachedDocument{}
	}
	if w.documentLocks == nil {
		w.documentLocks = map[string]*sync.Mutex{}
	}
	lock := w.documentLocks[target]
	if lock == nil {
		lock = &sync.Mutex{}
		w.documentLocks[target] = lock
	}
	w.documentMu.Unlock()
	lock.Lock()
	defer lock.Unlock()
	w.documentMu.Lock()
	old, ok := w.documents[target]
	w.documentMu.Unlock()
	if ok && time.Since(old.Document.ObservedAt) < 5*time.Minute {
		return old, nil
	}
	doc, e := w.Reader.Document(ctx, target)
	if e != nil {
		return CachedDocument{}, e
	}
	if len(doc.Text) > 64000 {
		doc.Text = string([]rune(doc.Text)[:min(len([]rune(doc.Text)), 16000)])
		doc.Truncated = true
	}
	origin := marketdata.Origin(doc.URL, doc.Text)
	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(doc.URL+":"+doc.ObservedAt.Format(time.RFC3339Nano)))
	out := CachedDocument{id, doc, origin, "read_source"}
	b, _ := json.Marshal(out)
	_, e = w.Store.Pool.Exec(ctx, "INSERT INTO evidence(id,origin_url,origin_key,source_url,published_at,observed_at,status,body) VALUES($1,$2,$3,$2,$4,$5,$6,$7)", id, doc.URL, origin, doc.PublishedAt, doc.ObservedAt, out.Status, b)
	if e == nil {
		w.documentMu.Lock()
		w.documents[target] = out
		w.documentMu.Unlock()
	}
	return out, e
}
func (w *Worker) Serve(ctx context.Context, listen string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /read-source", func(out http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(out, r.Body, 2048)
		var in struct {
			URL string `json:"url"`
		}
		d := json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		if e := d.Decode(&in); e != nil {
			http.Error(out, "invalid_request", 400)
			return
		}
		v, e := w.ReadDocument(r.Context(), in.URL)
		if e != nil {
			http.Error(out, e.Error(), 502)
			return
		}
		out.Header().Set("Content-Type", "application/json")
		json.NewEncoder(out).Encode(v)
	})
	srv := &http.Server{Addr: listen, Handler: mux, ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 50 * time.Second}
	go func() { <-ctx.Done(); srv.Shutdown(context.Background()) }()
	e := srv.ListenAndServe()
	if e == http.ErrServerClosed {
		return nil
	}
	return e
}
func CollectHTTP(ctx context.Context, base, target string) (CachedDocument, error) {
	var out CachedDocument
	b, _ := json.Marshal(map[string]string{"url": target})
	req, e := http.NewRequestWithContext(ctx, "POST", base+"/read-source", strings.NewReader(string(b)))
	if e != nil {
		return out, e
	}
	req.Header.Set("Content-Type", "application/json")
	res, e := (&http.Client{Timeout: 45 * time.Second}).Do(req)
	if e != nil {
		return out, e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return out, fmt.Errorf("source_read_failed_%d", res.StatusCode)
	}
	e = json.NewDecoder(http.MaxBytesReader(nil, res.Body, 128*1024)).Decode(&out)
	return out, e
}
