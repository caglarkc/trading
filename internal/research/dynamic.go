package research

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/contracts"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/modelrunner"
	"github.com/caglarkc/trading/internal/store"
	"github.com/caglarkc/trading/internal/worker"
	"github.com/google/uuid"
	"strings"
	"time"
)

func (s *Service) Dynamic(ctx context.Context, j jobs.Job) (any, error) {
	type candidate struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Exchange string `json:"exchange"`
		Reason   string `json:"reason"`
	}
	var existing int
	if e := s.Pool.QueryRow(ctx, "SELECT count(*) FROM universe_versions WHERE week_key=$1", j.Key).Scan(&existing); e != nil {
		return nil, e
	}
	if existing > 0 {
		return map[string]any{"unchanged": true}, nil
	}
	rows, e := s.Pool.Query(ctx, "SELECT symbol FROM assets WHERE fixed AND category='US' ORDER BY symbol")
	if e != nil {
		return nil, e
	}
	fixed := map[string]bool{}
	for rows.Next() {
		var sym string
		rows.Scan(&sym)
		fixed[sym] = true
	}
	rows.Close()
	fixedJSON, _ := json.Marshal(fixed)
	result, e := s.Runner.Run(ctx, modelrunner.Request{Wallet: "dynamic-" + j.ID.String(), Web: true, Schema: contracts.DynamicSchema(), Deadline: j.Deadline, Prompt: fmt.Sprintf("Select exactly 20 distinct US-listed individual common stocks outside this fixed list for the week beginning at %s UTC. Use current public listing and event/sector evidence; no ETFs, OTC, derivatives or remembered quotes. Balance actionable company events and sector coverage. Keep names/exchanges source-supported. This shared selection contains no account information. Fixed exclusions: %s. Return schema only.", time.Now().UTC().Format(time.RFC3339), fixedJSON)})
	if e != nil {
		return nil, e
	}
	var proposal struct {
		Candidates []candidate `json:"candidates"`
	}
	if e = json.Unmarshal(result.Output, &proposal); e != nil {
		return nil, e
	}
	if len(proposal.Candidates) != 20 {
		return nil, fmt.Errorf("dynamic_count_invalid")
	}
	seen := map[string]bool{}
	for _, c := range proposal.Candidates {
		if (c.Exchange != "NYSE" && c.Exchange != "NASDAQ") || fixed[c.Symbol] || seen[c.Symbol] {
			return nil, fmt.Errorf("dynamic_duplicate_or_fixed")
		}
		seen[c.Symbol] = true
		sa, e := worker.CollectHTTP(ctx, s.Collector, "https://stockanalysis.com/stocks/"+strings.ToLower(c.Symbol)+"/")
		if e != nil || !strings.Contains(strings.ToUpper(sa.Document.Title), "("+c.Symbol+")") || !strings.Contains(strings.ToUpper(sa.Document.Text), c.Exchange) {
			return nil, fmt.Errorf("dynamic_listing_unverified: %s", c.Symbol)
		}
		other, e := worker.CollectHTTP(ctx, s.Collector, "https://finance.yahoo.com/quote/"+strings.ReplaceAll(c.Symbol, ".", "-")+"/")
		if e != nil || !strings.Contains(strings.ToUpper(other.Document.Title), strings.ReplaceAll(c.Symbol, ".", "-")) {
			return nil, fmt.Errorf("dynamic_second_listing_unverified: %s", c.Symbol)
		}
	}
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if e = jobs.Guard(ctx, tx, j); e != nil {
		return nil, e
	}
	version := uuid.NewSHA1(j.ID, []byte("universe-v1"))
	body, _ := json.Marshal(proposal)
	_, e = tx.Exec(ctx, "INSERT INTO universe_versions(id,week_key,effective_at,body) VALUES($1,$2,now(),$3)", version, j.Key, body)
	if e != nil {
		return nil, e
	}
	for _, c := range proposal.Candidates {
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:asset:"+c.Symbol))
		aliases, _ := json.Marshal(map[string]string{"stockanalysis": "https://stockanalysis.com/stocks/" + strings.ToLower(c.Symbol) + "/", "yahoo": "https://finance.yahoo.com/quote/" + strings.ReplaceAll(c.Symbol, ".", "-") + "/", "webull": "https://www.webull.com/quote/" + strings.ToLower(c.Exchange) + "-" + strings.ToLower(strings.ReplaceAll(c.Symbol, ".", "-"))})
		_, e = tx.Exec(ctx, "INSERT INTO assets(id,symbol,name,category,exchange,aliases,fixed) VALUES($1,$2,$3,'US',$4,$5,false) ON CONFLICT(id) DO NOTHING", id, c.Symbol, c.Name, c.Exchange, aliases)
		if e != nil {
			return nil, e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO dynamic_members(version_id,asset_id) VALUES($1,$2)", version, id); e != nil {
			return nil, e
		}
	}
	if e = store.CancelRemovedBuys(ctx, tx, version); e != nil {
		return nil, e
	}
	if e = tx.Commit(ctx); e != nil {
		return nil, e
	}
	return map[string]any{"universe_version": version, "count": 20}, nil
}
