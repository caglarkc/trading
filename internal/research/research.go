package research

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/jobs"
	"github.com/caglarkc/trading/internal/worker"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Finding struct {
	Category string     `json:"category"`
	Symbols  []string   `json:"symbols"`
	Claim    string     `json:"claim"`
	Status   string     `json:"evidence_status"`
	Sources  []Source   `json:"sources"`
	Expected *time.Time `json:"expected_at"`
	Missing  string     `json:"missing"`
}
type Source struct {
	URL        string     `json:"url"`
	Excerpt    string     `json:"excerpt"`
	EvidenceID *uuid.UUID `json:"evidence_id"`
	Origin     string     `json:"origin_key"`
	Published  *time.Time `json:"published_at"`
	Observed   *time.Time `json:"observed_at"`
	Error      string     `json:"error,omitempty"`
}
type Bundle struct {
	AsOf        time.Time   `json:"as_of"`
	Findings    []Finding   `json:"findings"`
	EvidenceIDs []uuid.UUID `json:"evidence_ids"`
	Coverage    string      `json:"coverage"`
	Version     string      `json:"version"`
}
type Proposal struct {
	Findings []Finding `json:"findings"`
	Coverage string    `json:"coverage"`
}

func normalize(s string) string { return strings.Join(strings.Fields(s), " ") }
func Build(ctx context.Context, collector string, p Proposal) (Bundle, error) {
	b := Bundle{AsOf: time.Now().UTC(), Version: "research-v1", Coverage: p.Coverage, Findings: []Finding{}, EvidenceIDs: []uuid.UUID{}}
	if len(p.Findings) > 100 {
		return b, fmt.Errorf("research_too_large")
	}
	seen := map[uuid.UUID]bool{}
	for _, f := range p.Findings {
		if f.Category != "US" && f.Category != "QQQ" && f.Category != "TQQQ" && f.Category != "EWC" {
			return b, fmt.Errorf("invalid_research_category")
		}
		if len(f.Sources) > 4 {
			return b, fmt.Errorf("too_many_finding_sources")
		}
		origins := map[string]bool{}
		content := map[string]bool{}
		valid := 0
		expected := f.Expected
		f.Expected = nil
		for i := range f.Sources {
			s := &f.Sources[i]
			s.EvidenceID = nil
			s.Origin = ""
			s.Published = nil
			s.Observed = nil
			d, e := worker.CollectHTTP(ctx, collector, s.URL)
			if e != nil {
				s.Error = e.Error()
				continue
			}
			excerpt := normalize(s.Excerpt)
			if len(excerpt) < 20 || !strings.Contains(normalize(d.Document.Text), excerpt) {
				s.Error = "excerpt_not_in_read_content"
				continue
			}
			id := d.ID
			s.EvidenceID = &id
			s.Origin = d.Origin
			s.Published = d.Document.PublishedAt
			t := d.Document.ObservedAt
			s.Observed = &t
			fingerprint := marketHash(s.Excerpt)
			content[fingerprint] = true
			origins[s.Origin] = true
			valid++
			if !seen[id] {
				seen[id] = true
				b.EvidenceIDs = append(b.EvidenceIDs, id)
			}
			// A proposed schedule is kept only where an exact ISO event timestamp appears
			// in the read supporting text. Publication time is never event time.
			if expected != nil && strings.Contains(d.Document.Text, expected.Format(time.RFC3339)) {
				f.Expected = expected
			}
		}
		if f.Status != "conflicting" {
			switch {
			case valid == 0:
				f.Status = "missing"
			case len(origins) > 1 && len(content) > 1:
				f.Status = "two_sources"
			default:
				f.Status = "single_source"
			}
		}
		if expected != nil && f.Expected == nil {
			if f.Missing != "" {
				f.Missing += "; "
			}
			f.Missing += "event_time_unverified"
		}
		b.Findings = append(b.Findings, f)
	}
	return b, nil
}
func marketHash(s string) string { return normalize(strings.ToLower(s)) }
func Save(ctx context.Context, pool *pgxpool.Pool, j jobs.Job, b Bundle) error {
	tx, e := pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if e = jobs.Guard(ctx, tx, j); e != nil {
		return e
	}
	body, e := json.Marshal(b)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, "INSERT INTO research_bundles(id,as_of,body) VALUES($1,$2,$3) ON CONFLICT(id) DO NOTHING", j.ID, b.AsOf, body)
	if e != nil {
		return e
	}
	for _, f := range b.Findings {
		if f.Expected == nil || len(f.Sources) == 0 {
			continue
		}
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:event:"+f.Category+":"+f.Claim+":"+f.Expected.Format(time.RFC3339)))
		source := ""
		for _, s := range f.Sources {
			if s.EvidenceID != nil {
				source = s.URL
				break
			}
		}
		if source == "" {
			continue
		}
		fb, _ := json.Marshal(f)
		_, e = tx.Exec(ctx, "INSERT INTO calendar_events(id,category,event_key,expected_at,status,source_url,body) VALUES($1,$2,$3,$4,'verified_time',$5,$6) ON CONFLICT(event_key) DO NOTHING", id, f.Category, id.String(), f.Expected, source, fb)
		if e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
