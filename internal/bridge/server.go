package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/httpapi"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/store"
	"github.com/caglarkc/trading/internal/worker"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Scope struct {
	Wallet, Decision string
	Job              jobs.Job
	Secret           []byte
	API, Collector   string
	Store            *store.Store
}

func Start(ctx context.Context, socket string, s Scope) (func(), error) {
	if e := os.MkdirAll(filepath.Dir(socket), 0700); e != nil {
		return nil, e
	}
	if old, e := os.Lstat(socket); e == nil {
		if old.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("bridge_path_not_socket")
		}
		os.Remove(socket)
	}
	l, e := net.Listen("unix", socket)
	if e != nil {
		return nil, e
	}
	if e = os.Chmod(socket, 0600); e != nil {
		l.Close()
		return nil, e
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tool", s.tool)
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	go srv.Serve(l)
	stop := func() { srv.Close(); os.Remove(socket) }
	go func() { <-ctx.Done(); stop() }()
	return stop, nil
}
func (s Scope) tool(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	var in struct {
		Name      string          `json:"Name"`
		Arguments json.RawMessage `json:"Arguments"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(&in); e != nil {
		http.Error(w, "invalid_tool_request", 400)
		return
	}
	if in.Name == "read_source" {
		var args struct {
			URL string `json:"url"`
		}
		dec := json.NewDecoder(bytes.NewReader(in.Arguments))
		dec.DisallowUnknownFields()
		if e := dec.Decode(&args); e != nil || len(args.URL) > 1800 {
			http.Error(w, "invalid_source_request", 400)
			return
		}
		doc, e := worker.CollectHTTP(r.Context(), s.Collector, args.URL)
		if e != nil {
			http.Error(w, e.Error(), 502)
			return
		}
		tx, e := s.Store.Pool.Begin(r.Context())
		if e != nil {
			http.Error(w, "evidence_record_failed", 500)
			return
		}
		defer tx.Rollback(r.Context())
		if e = jobs.Guard(r.Context(), tx, s.Job); e != nil {
			http.Error(w, "lease_lost", 409)
			return
		}
		_, e = tx.Exec(r.Context(), "INSERT INTO decision_evidence(wallet_id,decision_id,evidence_id) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", s.Wallet, s.Decision, doc.ID)
		if e != nil {
			http.Error(w, "evidence_record_failed", 500)
			return
		}
		if e = tx.Commit(r.Context()); e != nil {
			http.Error(w, "evidence_record_failed", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
		return
	}
	routes := map[string]string{"account": "account", "positions": "positions", "orders": "orders", "receipts": "receipts", "memory": "memory", "evidence": "evidence"}
	route, ok := routes[in.Name]
	if !ok {
		http.Error(w, "tool_not_allowed", 403)
		return
	}
	var args struct {
		Cursor int64 `json:"cursor"`
		Limit  int   `json:"limit"`
	}
	if len(in.Arguments) == 0 {
		in.Arguments = []byte(`{}`)
	}
	dec := json.NewDecoder(bytes.NewReader(in.Arguments))
	dec.DisallowUnknownFields()
	if e := dec.Decode(&args); e != nil || args.Cursor < 0 || args.Limit < 0 || args.Limit > 200 {
		http.Error(w, "invalid_read_request", 400)
		return
	}
	if args.Limit == 0 {
		args.Limit = 50
	}
	token := httpapi.Sign(s.Secret, httpapi.Claims{Wallet: s.Wallet, Decision: s.Decision, Job: s.Job.ID.String(), Fence: s.Job.Fence, Expires: time.Now().Add(5 * time.Minute).Unix()})
	req, e := http.NewRequestWithContext(r.Context(), "GET", s.API+"/agent/"+route+"?cursor="+strconv.FormatInt(args.Cursor, 10)+"&limit="+strconv.Itoa(args.Limit), nil)
	if e != nil {
		http.Error(w, "account_read_failed", 500)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, e := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if e != nil {
		http.Error(w, "account_unavailable", 502)
		return
	}
	defer res.Body.Close()
	b, e := io.ReadAll(io.LimitReader(res.Body, 256*1024+1))
	if e != nil || len(b) > 256*1024 {
		http.Error(w, "account_result_too_large", 502)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	w.Write(b)
}
func Submit(ctx context.Context, s Scope, action any) (json.RawMessage, error) {
	b, e := json.Marshal(action)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", s.API+"/agent/order-actions", bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+httpapi.Sign(s.Secret, httpapi.Claims{Wallet: s.Wallet, Decision: s.Decision, Job: s.Job.ID.String(), Fence: s.Job.Fence, Expires: time.Now().Add(5 * time.Minute).Unix()}))
	res, e := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	out, e := io.ReadAll(io.LimitReader(res.Body, 256*1024+1))
	if e != nil {
		return nil, e
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("order_action_http_%d", res.StatusCode)
	}
	return json.RawMessage(out), nil
}

var _ = uuid.Nil
