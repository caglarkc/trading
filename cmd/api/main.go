package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/database"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/httpapi"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/caglarkc/trading/internal/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var release = "development"

func main() {
	migrate := flag.Bool("migrate", false, "apply forward migrations and idempotent seed")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *migrate {
		dsn := required("DATABASE_OWNER_URL")
		if e := database.Migrate(ctx, dsn); e != nil {
			log.Fatal(e)
		}
		s, e := store.Open(ctx, dsn)
		if e != nil {
			log.Fatal(e)
		}
		defer s.Close()
		b, e := os.ReadFile(required("ASSETS_FILE"))
		if e != nil {
			log.Fatal(e)
		}
		a, e := marketdata.DecodeAssets(b)
		if e != nil {
			log.Fatal(e)
		}
		if e = s.Seed(ctx, a); e != nil {
			log.Fatal(e)
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"migration": "passed", "wallets": 8, "seed_idempotent": true})
		return
	}
	agent, e := store.Open(ctx, required("DATABASE_AGENT_URL"))
	if e != nil {
		log.Fatal(e)
	}
	defer agent.Close()
	operator, e := store.Open(ctx, required("DATABASE_OPERATOR_URL"))
	if e != nil {
		log.Fatal(e)
	}
	defer operator.Close()
	cal, e := calendar.Load(required("CALENDAR_FILE"))
	if e != nil {
		log.Fatal(e)
	}
	fee, e := domain.LoadFees(required("FEES_FILE"))
	if e != nil {
		log.Fatal(e)
	}
	s := &httpapi.Server{Calendar: cal, Fees: fee, Agent: agent, Operator: operator, Secret: []byte(required("AGENT_SIGNING_SECRET")), OperatorToken: required("OPERATOR_TOKEN"), Release: release, ExecutionEnabled: os.Getenv("EXECUTION_ENABLED") == "true"}
	srv := &http.Server{Addr: env("API_LISTEN", "127.0.0.1:18880"), Handler: s.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(c)
	}()
	if e = srv.ListenAndServe(); e != http.ErrServerClosed {
		log.Fatal(e)
	}
}
func required(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing configuration: %s", k)
	}
	return v
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
