package marketdata

import (
	"strings"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }
func testRows(now time.Time) []Observation {
	return []Observation{
		{ID: "one", AssetID: "asset", SourceID: "one", IdentityVerified: true, Price: ptr("99.5"), Currency: "USD", PriceType: "LAST", Session: "REGULAR", ObservedAt: now.Add(-time.Minute)},
		{ID: "two", AssetID: "asset", SourceID: "two", IdentityVerified: true, Price: ptr("100"), Currency: "USD", PriceType: "LAST", Session: "REGULAR", ObservedAt: now.Add(-30 * time.Second)},
		{ID: "three", AssetID: "asset", SourceID: "three", IdentityVerified: true, Price: ptr("100.5"), Currency: "USD", PriceType: "LAST", Session: "REGULAR", ObservedAt: now},
	}
}
func TestPriceBoundaryAndEvidenceFailures(t *testing.T) {
	now := time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC)
	if v := Validate("asset", testRows(now), now, true); !v.Verified || *v.RangePercent != "1" {
		t.Fatal(v)
	}
	cases := []struct {
		name   string
		change func([]Observation) []Observation
	}{
		{"over_one_percent", func(r []Observation) []Observation { r[2].Price = ptr("100.500001"); return r }},
		{"stale", func(r []Observation) []Observation {
			r[0].ObservedAt = now.Add(-time.Minute - time.Nanosecond)
			return r
		}},
		{"future", func(r []Observation) []Observation { r[2].ObservedAt = now.Add(time.Nanosecond); return r }},
		{"missing_read_time", func(r []Observation) []Observation { r[0].ObservedAt = time.Time{}; return r }},
		{"currency", func(r []Observation) []Observation { r[0].Currency = "CAD"; return r }},
		{"duplicate_source", func(r []Observation) []Observation { r[0].SourceID = r[1].SourceID; return r }},
		{"wrong_share_class", func(r []Observation) []Observation { r[0].AssetID = "other-class"; return r }},
		{"two_sources", func(r []Observation) []Observation { return r[:2] }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if Validate("asset", c.change(testRows(now)), now, true).Verified {
				t.Fatal("unsafe quote accepted")
			}
		})
	}
	if Validate("asset", testRows(now), now, false).Verified {
		t.Fatal("market closed")
	}
}
func TestBBOMissingTimeOrUnits(t *testing.T) {
	now := time.Now().UTC()
	r := Observation{IdentityVerified: true, Currency: "USD", Session: "REGULAR", Bid: ptr("99"), Ask: ptr("100"), BidSize: ptr(int64(100)), AskSize: ptr(int64(100)), BBOTime: ptr(now), SizeUnit: "shares", VenueScope: "NASDAQ_BASIC"}
	if !VerifyLiquidity(r, now) {
		t.Fatal("valid BBO rejected")
	}
	r.SizeUnit = ""
	if VerifyLiquidity(r, now) {
		t.Fatal("unit invented")
	}
	r.SizeUnit = "shares"
	r.BBOTime = nil
	if VerifyLiquidity(r, now) {
		t.Fatal("time invented")
	}
}

func TestReadTimeValidationDoesNotRequireVendorMetadata(t *testing.T) {
	now := time.Now().UTC()
	rows := testRows(now)
	for i := range rows {
		rows[i].EventTime = nil
		rows[i].Currency = ""
		rows[i].Session = "UNKNOWN"
		rows[i].PriceType = "UNKNOWN"
		rows[i].Delayed = true
	}
	if v := Validate("asset", rows, now, true); !v.Verified {
		t.Fatal(v)
	}
	rows[0].ObservedAt = now.Add(-61 * time.Second)
	rows[0].EventTime = &now
	if v := Validate("asset", rows, now, true); v.Verified {
		t.Fatal("vendor timestamp hid a stale read", v)
	}
}
func TestHTMLAndClosedTimeAttribution(t *testing.T) {
	body := `<html><title>QQQ Price</title><body><div>USD</div><h1>Invesco QQQ (QQQ)</h1><div class="text-4xl">747.58</div><span>At close:</span><span>Oct 9, 2026, 4:00 PM EDT</span><div>After-hours: 750.00</div><script>price=9999; token="private"</script></body></html>`
	p := HTMLPage(body)
	if strings.Contains(p.Text, "token") || strings.Contains(p.Text, "9999") {
		t.Fatal("script leaked into visible source")
	}
	r := Parse("stockanalysis", Asset{ID: "QQQ", Symbol: "QQQ"}, "https://stockanalysis.com/etf/qqq/", p, time.Now())
	if r.Price == nil || *r.Price != "747.58" || r.PriceType != "CLOSE" || r.EventTime == nil || r.EventTime.Hour() != 20 {
		t.Fatal(r)
	}
	w := Parse("webull", Asset{ID: "EWC", Symbol: "EWC"}, "url", Page{Title: "EWC", Heading: "EWC", Text: "EWC\n58.32\nClosed 19:30 10/09 EDT"}, time.Now())
	if w.EventTime != nil || w.Session != "UNKNOWN" {
		t.Fatal("generic closed clock was used as regular time")
	}
}
func TestReaderRejectsPrivateTargets(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "localhost", "169.254.169.254", "::1"} {
		if _, e := safeIPs(t.Context(), host); e == nil {
			t.Fatal("private destination permitted", host)
		}
	}
}
