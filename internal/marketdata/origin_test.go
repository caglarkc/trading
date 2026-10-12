package marketdata

import (
	"context"
	"testing"
	"time"
)

func TestSourceBoundariesAndProvenance(t *testing.T) {
	r := NewReader("")
	for _, u := range []string{"http://stockanalysis.com", "https://stockanalysis.com:444/", "https://stockanalysis.com.attacker.test/", "https://user:pass@stockanalysis.com/", "https://127.0.0.1/"} {
		if r.permitted(u) {
			t.Fatal("invalid source allowed", u)
		}
	}
	for _, host := range []string{"100.100.100.200", "198.18.0.1", "2001:db8::1"} {
		if _, e := safeIPs(context.Background(), host); e == nil {
			t.Fatal("reserved address allowed", host)
		}
	}
	if Origin("https://site-one.example/article", "Via Reuters") != Origin("https://site-two.example/article", "Reported by Reuters") {
		t.Fatal("wire copies counted as independent")
	}
	p := HTMLPage(`<html><head><meta property="article:published_time" content="2026-10-09T14:00:00-04:00"></head><body>Visible source text</body></html>`)
	if p.PublishedAt == nil || !p.PublishedAt.Equal(time.Date(2026, 10, 9, 18, 0, 0, 0, time.UTC)) {
		t.Fatal(p)
	}
	p = HTMLPage(`<html><body>Published yesterday</body></html>`)
	if p.PublishedAt != nil {
		t.Fatal("invented publication time")
	}
}
