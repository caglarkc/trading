package research

import (
	"context"
	"encoding/json"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/caglarkc/trading/internal/worker"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOriginsExcerptsAndEventTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			URL string `json:"url"`
		}
		json.NewDecoder(r.Body).Decode(&in)
		origin := "reuters"
		text := "Issuer reports annual audited results in this exact excerpt."
		if in.URL == "second" {
			origin = "issuer"
			text = "The second independent source confirms the scheduled earnings date."
		}
		json.NewEncoder(w).Encode(worker.CachedDocument{ID: uuid.NewSHA1(uuid.NameSpaceURL, []byte(in.URL)), Origin: origin, Document: marketdata.Document{Text: text, ObservedAt: now}})
	}))
	defer server.Close()
	p := Proposal{Findings: []Finding{
		{Category: "US", Expected: &now, Sources: []Source{{URL: "first", Excerpt: "Issuer reports annual audited results"}, {URL: "copy", Excerpt: "Issuer reports annual audited results"}}},
		{Category: "QQQ", Sources: []Source{{URL: "first", Excerpt: "Issuer reports annual audited results"}, {URL: "second", Excerpt: "The second independent source confirms"}}},
		{Category: "EWC", Sources: []Source{{URL: "first", Excerpt: "A made up excerpt without actual support"}}},
	}}
	b, e := Build(context.Background(), server.URL, p)
	if e != nil {
		t.Fatal(e)
	}
	if b.Findings[0].Status != "single_source" || b.Findings[0].Expected != nil || b.Findings[0].Missing != "event_time_unverified" {
		t.Fatal(b.Findings[0])
	}
	if b.Findings[1].Status != "two_sources" || b.Findings[2].Status != "missing" || b.Findings[2].Sources[0].EvidenceID != nil {
		t.Fatal(b)
	}
}
