package httpapi

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type contextKey int

const claimsKey contextKey = 1

type Server struct {
	Agent            *store.Store
	Operator         *store.Store
	Secret           []byte
	OperatorToken    string
	Release          string
	Calendar         calendar.Calendar
	Fees             domain.FeeProfile
	ExecutionEnabled bool
}

func ClaimsFrom(ctx context.Context) Claims { return ctx.Value(claimsKey).(Claims) }
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code string) {
	respond(w, status, map[string]string{"error": code})
}
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, c := context.WithTimeout(r.Context(), 2*time.Second)
		defer c()
		if e := s.Operator.Pool.Ping(ctx); e != nil {
			fail(w, 503, "db_unavailable")
			return
		}
		respond(w, 200, map[string]any{"status": "ready", "release": s.Release, "execution_enabled": s.ExecutionEnabled})
	})
	r.Route("/agent", func(r chi.Router) {
		r.Use(s.agentAuth)
		r.Post("/order-actions", s.orderAction)
		r.Get("/account", s.account)
		r.Get("/positions", s.account)
		r.Get("/orders", s.agentRows("orders"))
		r.Get("/receipts", s.agentRows("receipts"))
		r.Get("/memory", func(w http.ResponseWriter, r *http.Request) {
			v, e := s.Agent.LatestMemory(r.Context(), ClaimsFrom(r.Context()).Wallet)
			if e != nil {
				fail(w, 500, "read_failed")
				return
			}
			respond(w, 200, v)
		})
		r.Get("/evidence", s.sharedEvidence)
	})
	r.Route("/operator", func(r chi.Router) {
		r.Use(s.operatorAuth)
		r.Get("/observation", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
			defer cancel()
			v, e := s.Operator.Observation(ctx)
			if e != nil {
				fail(w, 503, "observation_unavailable")
				return
			}
			v.Execution = v.Execution && s.ExecutionEnabled
			respond(w, 200, v)
		})
		r.Get("/wallets", func(w http.ResponseWriter, r *http.Request) {
			v, e := s.Operator.Wallets(r.Context())
			if e != nil {
				fail(w, 500, "read_failed")
				return
			}
			respond(w, 200, v)
		})
		r.Get("/wallets/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if _, e := uuid.Parse(id); e != nil {
				fail(w, 400, "invalid_wallet")
				return
			}
			v, e := s.Operator.Wallet(r.Context(), id)
			if e != nil {
				fail(w, 404, "wallet_not_found")
				return
			}
			respond(w, 200, v)
		})
		r.Get("/orders", s.operatorRows("orders"))
		r.Get("/evidence", s.operatorRows("evidence"))
		r.Get("/system-status", func(w http.ResponseWriter, r *http.Request) {
			v, e := s.Operator.Rows(r.Context(), "events", "", 0, 50)
			if e != nil {
				fail(w, 500, "read_failed")
				return
			}
			respond(w, 200, map[string]any{"as_of": time.Now().UTC(), "release": s.Release, "execution_enabled": s.ExecutionEnabled, "events": v})
		})
	})
	return r
}
func (s *Server) agentAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		c, e := Verify(s.Secret, token, time.Now())
		if e != nil {
			fail(w, 401, "invalid_agent_scope")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, c)))
	})
}
func (s *Server) operatorAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if s.OperatorToken == "" || !EqualSecret(token, s.OperatorToken) {
			fail(w, 401, "operator_auth_required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) account(w http.ResponseWriter, r *http.Request) {
	v, e := s.Agent.Wallet(r.Context(), ClaimsFrom(r.Context()).Wallet)
	if e != nil {
		fail(w, 404, "wallet_not_found")
		return
	}
	respond(w, 200, v)
}
func pageArgs(r *http.Request) (int64, int) {
	cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
	if cursor < 0 {
		cursor = 0
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	return cursor, limit
}
func (s *Server) agentRows(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, l := pageArgs(r)
		v, e := s.Agent.Rows(r.Context(), kind, ClaimsFrom(r.Context()).Wallet, c, l)
		if e != nil {
			fail(w, 500, "read_failed")
			return
		}
		respond(w, 200, v)
	}
}
func (s *Server) operatorRows(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, l := pageArgs(r)
		v, e := s.Operator.Rows(r.Context(), kind, "", c, l)
		if e != nil {
			fail(w, 500, "read_failed")
			return
		}
		respond(w, 200, v)
	}
}
func (s *Server) sharedEvidence(w http.ResponseWriter, r *http.Request) {
	c, l := pageArgs(r)
	v, e := s.Agent.Rows(r.Context(), "evidence", "", c, l)
	if e != nil {
		fail(w, 500, "read_failed")
		return
	}
	respond(w, 200, v)
}
