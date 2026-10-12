package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/database/dbgen"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *dbgen.Queries
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	c, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, e
	}
	c.MaxConns = 4
	c.MinConns = 0
	c.ConnConfig.RuntimeParams["statement_timeout"] = "10000"
	c.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = "15000"
	p, e := pgxpool.NewWithConfig(ctx, c)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, e
	}
	return &Store{p, dbgen.New(p)}, nil
}
func (s *Store) Close() { s.Pool.Close() }
func (s *Store) Scoped(ctx context.Context, wallet string, fn func(pgx.Tx, *dbgen.Queries) error) error {
	id, e := uuid.Parse(wallet)
	if e != nil {
		return fmt.Errorf("invalid_wallet")
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT set_config('app.wallet_id',$1,true)", id.String()); e != nil {
		return e
	}
	if e = fn(tx, s.Queries.WithTx(tx)); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

type Wallet struct {
	ID       uuid.UUID      `json:"id"`
	Code     string         `json:"code"`
	Category string         `json:"category"`
	Version  int64          `json:"account_version"`
	Account  domain.Account `json:"account"`
	AsOf     time.Time      `json:"as_of"`
}

func DecodeWallet(row dbgen.Wallet) (Wallet, error) {
	var a domain.Account
	e := json.Unmarshal(row.State, &a)
	if a.DividendIncome == "" {
		a.DividendIncome = "0"
	}
	if a.Receivables == nil {
		a.Receivables = []domain.Receivable{}
	}
	return Wallet{row.ID, row.Code, row.Category, row.AccountVersion, a, time.Now().UTC()}, e
}
func (s *Store) Wallet(ctx context.Context, id string) (Wallet, error) {
	var out Wallet
	e := s.Scoped(ctx, id, func(_ pgx.Tx, q *dbgen.Queries) error {
		r, e := q.GetWallet(ctx, uuid.MustParse(id))
		if e != nil {
			return e
		}
		out, e = DecodeWallet(r)
		return e
	})
	return out, e
}
func (s *Store) Wallets(ctx context.Context) ([]Wallet, error) {
	rows, e := s.Queries.ListWallets(ctx)
	if e != nil {
		return nil, e
	}
	out := []Wallet{}
	for _, r := range rows {
		w, e := DecodeWallet(r)
		if e != nil {
			return nil, e
		}
		out = append(out, w)
	}
	return out, nil
}

var Codes = []string{"US-FAST", "US-EVENT", "US-PATIENT", "QQQ-FAST", "QQQ-PATIENT", "TQQQ-FAST", "TQQQ-PATIENT", "EWC-BALANCED"}

func WalletID(code string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:wallet:"+code))
}
func (s *Store) Seed(ctx context.Context, assets []marketdata.Asset) error {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	for _, a := range assets {
		aliases, _ := json.Marshal(a.Aliases)
		_, e = tx.Exec(ctx, "INSERT INTO assets(id,symbol,name,category,exchange,aliases) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO NOTHING", a.ID, a.Symbol, a.Name, a.Category, a.Exchange, aliases)
		if e != nil {
			return e
		}
	}
	state, _ := json.Marshal(domain.NewAccount())
	for i, code := range Codes {
		cat := "US"
		if i > 2 {
			cat = "QQQ"
			if i > 4 {
				cat = "TQQQ"
			}
			if i == 7 {
				cat = "EWC"
			}
		}
		id := WalletID(code)
		tag, e := tx.Exec(ctx, "INSERT INTO wallets(id,code,category,state) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO NOTHING", id, code, cat, state)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 1 {
			jid := uuid.New()
			_, e = tx.Exec(ctx, "INSERT INTO journal_transactions(id,wallet_id,event_type,external_event_id,occurred_at) VALUES($1,$2,'initial_capital','initial-v1',now())", jid, id)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, "INSERT INTO journal_lines(wallet_id,transaction_id,account_code,amount,currency) VALUES($1,$2,'cash_available',100000,'USD'),($1,$2,'capital',-100000,'USD')", id, jid)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, "INSERT INTO wallet_sessions(wallet_id) VALUES($1)", id)
			if e != nil {
				return e
			}
		}
	}
	return tx.Commit(ctx)
}
