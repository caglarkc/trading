package domain

import (
	"encoding/json"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"os"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }
func fixtures(t *testing.T) (calendar.Calendar, FeeProfile, time.Time, string) {
	t.Helper()
	cal, e := calendar.Load("../../config/calendar.json")
	if e != nil {
		t.Fatal(e)
	}
	var f FeeProfile
	b, e := os.ReadFile("../../config/fees.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	return cal, f, time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC), uuid.NewString()
}
func quote(asset string, now time.Time, last, bid, ask string, size int64) marketdata.Snapshot {
	rows := []marketdata.Observation{}
	for _, source := range []string{"stockanalysis", "yahoo", "webull"} {
		rows = append(rows, marketdata.Observation{ID: uuid.NewString(), AssetID: asset, SourceID: source, IdentityVerified: true, Currency: "USD", PriceType: "LAST", Session: "REGULAR", EventTime: ptr(now), ObservedAt: now, Price: ptr(last), Bid: ptr(bid), Ask: ptr(ask), BidSize: ptr(size), AskSize: ptr(size), BBOTime: ptr(now), SizeUnit: "shares", VenueScope: "fixture-venue"})
	}
	return marketdata.Snapshot{Asset: marketdata.Asset{ID: asset}, Observations: rows}
}
func buy(asset string, q int64, budget string) Action {
	return Action{Action: "SUBMIT", AssetID: ptr(asset), Side: ptr("BUY"), Quantity: ptr(q), Type: ptr("MARKET"), TIF: ptr("DAY"), Budget: ptr(budget)}
}
func TestPartialCapacityFIFOAndSettlement(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	id := uuid.NewString()
	if e := a.Submit(buy(asset, 10, "2000"), id, now, cal, fee, true); e != nil {
		t.Fatal(e)
	}
	s := quote(asset, now.Add(time.Second), "100", "99.99", "100", 4)
	f, e := a.TryFill(id, s, now.Add(time.Second), cal, fee)
	if e != nil || f.Quantity != 4 || f.Price != "100.05" {
		t.Fatal(f, e)
	}
	if _, e = a.TryFill(id, s, now.Add(2*time.Second), cal, fee); e == nil {
		t.Fatal("same liquidity reused")
	}
	s = quote(asset, now.Add(3*time.Second), "100", "99.99", "100", 20)
	f, e = a.TryFill(id, s, now.Add(3*time.Second), cal, fee)
	if e != nil || f.Quantity != 6 || f.Commission != "0" || a.Orders[id].State != "filled" || a.Reserved != "0" {
		t.Fatal(f, e, a)
	}
	sell := Action{AssetID: ptr(asset), Side: ptr("SELL"), Quantity: ptr(int64(5)), Type: ptr("MARKET"), TIF: ptr("DAY")}
	sid := uuid.NewString()
	if e = a.Submit(sell, sid, now.Add(4*time.Second), cal, fee, false); e != nil {
		t.Fatal(e)
	}
	s = quote(asset, now.Add(5*time.Second), "110", "110", "110.01", 20)
	f, e = a.TryFill(sid, s, now.Add(5*time.Second), cal, fee)
	if e != nil || a.Shares(asset) != 5 || a.Unsettled == "0" {
		t.Fatal(f, e, a)
	}
	cash := a.Available
	if a.Settle(now.Add(time.Hour)) || cash != a.Available {
		t.Fatal("early settlement")
	}
	due := a.Settlements[0].Due
	if due.In(calendar.ET).Format("2006-01-02") != "2026-10-13" {
		t.Fatal(due)
	}
	if !a.Settle(due) || a.Unsettled != "0" || a.Settle(due) {
		t.Fatal("settlement replay", a)
	}
	if e = a.Check(); e != nil {
		t.Fatal(e)
	}
}
func TestOCOStopAndCancel(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	a.Lots = []Lot{{ID: uuid.NewString(), AssetID: asset, Quantity: 10, Cost: "1000", AcquiredAt: now.Add(-time.Hour)}}
	group := uuid.NewString()
	limit := Action{AssetID: ptr(asset), Side: ptr("SELL"), Quantity: ptr(int64(10)), Type: ptr("LIMIT"), TIF: ptr("DAY"), Limit: ptr("105"), OCO: &group}
	stop := limit
	stop.Type = ptr("STOP_MARKET")
	stop.Limit = nil
	stop.Stop = ptr("95")
	l, sid := uuid.NewString(), uuid.NewString()
	if e := a.Submit(limit, l, now, cal, fee, false); e != nil {
		t.Fatal(e)
	}
	if e := a.Submit(stop, sid, now, cal, fee, false); e != nil {
		t.Fatal(e)
	}
	if a.ReservedShares(asset, "") != 10 {
		t.Fatal("OCO double reservation")
	}
	plain := limit
	plain.OCO = nil
	if e := a.Submit(plain, uuid.NewString(), now, cal, fee, false); e == nil {
		t.Fatal("oversell accepted")
	}
	s := quote(asset, now.Add(time.Second), "106", "106", "106.01", 4)
	f, e := a.TryFill(l, s, now.Add(time.Second), cal, fee)
	if e != nil || f.Quantity != 4 || a.Orders[sid].Remaining != 6 {
		t.Fatal(f, e)
	}
	s = quote(asset, now.Add(2*time.Second), "90", "90", "90.01", 6)
	f, e = a.TryFill(sid, s, now.Add(2*time.Second), cal, fee)
	if e != nil || f.Price == "95" || a.Orders[l].State != "cancelled" || a.Shares(asset) != 0 {
		t.Fatal(f, e, a)
	}
}
func TestDataAndRiskGuards(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	if e := a.Submit(buy(asset, 100, "100001"), uuid.NewString(), now, cal, fee, true); e == nil {
		t.Fatal("debt allowed")
	}
	id := uuid.NewString()
	if e := a.Submit(buy(asset, 10, "1000"), id, now, cal, fee, true); e != nil {
		t.Fatal(e)
	}
	s := quote(asset, now.Add(time.Second), "100", "99.99", "100", 20)
	s.Observations[0].BidSize = nil
	if _, e := a.TryFill(id, s, now.Add(time.Second), cal, fee); e != nil {
		t.Fatal("other verified BBO can be used", e)
	}
	if a.Orders[id].Filled >= 10 {
		t.Fatal("cash exceeded")
	}
	id2 := uuid.NewString()
	if e := a.Submit(buy(asset, 1, "500"), id2, now.Add(2*time.Second), cal, fee, true); e != nil {
		t.Fatal(e)
	}
	s = quote(asset, now.Add(3*time.Second), "100", "99.99", "100", 20)
	for i := range s.Observations {
		s.Observations[i].BBOTime = nil
	}
	if _, e := a.TryFill(id2, s, now.Add(3*time.Second), cal, fee); e == nil {
		t.Fatal("unknown BBO filled")
	}
	a.Expire(now.Add(124 * time.Second))
	if a.Orders[id2].State != "expired" || a.Reserved != "0" {
		t.Fatal(a)
	}
}
func TestCalendarDSTHolidayAndEarlyClose(t *testing.T) {
	cal, _, _, _ := fixtures(t)
	fri := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	due, e := cal.Settlement(fri)
	if e != nil || due.In(calendar.ET).Format("2006-01-02") != "2026-10-13" {
		t.Fatal(due, e)
	}
	for _, tc := range []struct {
		date        string
		open, close string
	}{{"2026-10-12", "13:30", "20:00"}, {"2026-11-02", "14:30", "21:00"}, {"2026-11-27", "14:30", "18:00"}} {
		d, _ := time.Parse("2006-01-02", tc.date)
		o, c, day, e := cal.Session(d.Add(16 * time.Hour))
		if e != nil || !day || o.Format("15:04") != tc.open || c.Format("15:04") != tc.close {
			t.Fatal(tc, o, c, e)
		}
	}
	if cal.Open(time.Date(2026, 11, 26, 16, 0, 0, 0, time.UTC)) {
		t.Fatal("holiday open")
	}
}
func TestLimitStopLimitAndCorporate(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	x := buy(asset, 10, "1100")
	x.Type = ptr("STOP_LIMIT")
	x.Stop = ptr("101")
	x.Limit = ptr("102")
	id := uuid.NewString()
	if e := a.Submit(x, id, now, cal, fee, true); e != nil {
		t.Fatal(e)
	}
	s := quote(asset, now.Add(time.Second), "100", "100", "100.01", 20)
	if _, e := a.TryFill(id, s, now.Add(time.Second), cal, fee); e == nil || a.Orders[id].Triggered {
		t.Fatal("early trigger")
	}
	s = quote(asset, now.Add(2*time.Second), "103", "103", "103.01", 20)
	if _, e := a.TryFill(id, s, now.Add(2*time.Second), cal, fee); e == nil || !a.Orders[id].Triggered {
		t.Fatal("limit/trigger missing")
	}
	s = quote(asset, now.Add(3*time.Second), "101", "101", "101.01", 20)
	if _, e := a.TryFill(id, s, now.Add(3*time.Second), cal, fee); e != nil {
		t.Fatal(e)
	}
	split := CorporateEvent{ID: uuid.NewString(), AssetID: asset, Kind: "SPLIT", Verified: true, Ratio: "2", EffectiveAt: now}
	cost := a.Lots[0].Cost
	if e := a.Split(split); e != nil || a.Shares(asset) != 20 || a.Lots[0].Cost != cost {
		t.Fatal(e, a)
	}
	if e := a.Split(split); e != nil || a.Shares(asset) != 20 {
		t.Fatal("split replay")
	}
	reverse := split
	reverse.ID = uuid.NewString()
	reverse.Ratio = "0.03"
	if e := a.Split(reverse); e == nil || a.Shares(asset) != 20 {
		t.Fatal("fractional guessed")
	}
	div := CorporateEvent{ID: uuid.NewString(), AssetID: asset, Kind: "DIVIDEND", Verified: true, PerShare: "1", Classification: "ordinary_us", EffectiveAt: now, PayAt: now.AddDate(0, 0, 2)}
	if e := a.DividendEntitlement(div, 20); e != nil || a.Dividend != "14" {
		t.Fatal(e, a)
	}
	if e := a.DividendEntitlement(div, 20); e != nil || a.Dividend != "14" {
		t.Fatal("dividend replay")
	}
	cash := a.Available
	if a.PayDividends(now) || a.Available != cash {
		t.Fatal("early payment")
	}
	if !a.PayDividends(div.PayAt) || a.Dividend != "0" || a.PayDividends(div.PayAt) {
		t.Fatal("payment replay")
	}
	unknown := div
	unknown.ID = uuid.NewString()
	unknown.Classification = "ric_unclassified"
	if e := a.DividendEntitlement(unknown, 20); e != nil {
		t.Fatal(e)
	}
	cash = a.Available
	if a.PayDividends(unknown.PayAt) || cash != a.Available || a.Dividend != "20" {
		t.Fatal("unclassified tax/cash guessed")
	}
}
func TestProvenMinuteVolumeCapacity(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	id := uuid.NewString()
	if e := a.Submit(buy(asset, 20, "3000"), id, now, cal, fee, true); e != nil {
		t.Fatal(e)
	}
	at := now.Add(time.Second)
	s := quote(asset, at, "100", "100", "100", 100)
	for i := range s.Observations {
		s.Observations[i].MinuteVolume = ptr(int64(500))
		s.Observations[i].VolumeMinute = &at
		s.Observations[i].VolumeScope = "fixture-venue"
	}
	f, e := a.TryFill(id, s, at, cal, fee)
	if e != nil || f.Quantity != 5 {
		t.Fatal(f, e)
	}
	for i := range s.Observations {
		s.Observations[i].BBOTime = ptr(now.Add(2 * time.Second))
		s.Observations[i].EventTime = ptr(now.Add(2 * time.Second))
		s.Observations[i].ObservedAt = now.Add(2 * time.Second)
	}
	if _, e = a.TryFill(id, s, now.Add(2*time.Second), cal, fee); e == nil {
		t.Fatal("minute volume reused by new quote")
	}
}

func TestObservedPriceFillWithoutLiquidity(t *testing.T) {
	cal, fee, now, asset := fixtures(t)
	a := NewAccount()
	id := uuid.NewString()
	if e := a.Submit(buy(asset, 10, "2000"), id, now, cal, fee, true); e != nil {
		t.Fatal(e)
	}
	at := now.Add(time.Second)
	s := quote(asset, at, "100", "1", "999", 1)
	s.ExecutionMode = marketdata.ObservedPriceMode
	for i := range s.Observations {
		s.Observations[i].EventTime = nil
		s.Observations[i].BBOTime = nil
		s.Observations[i].Bid = nil
		s.Observations[i].Ask = nil
		s.Observations[i].BidSize = nil
		s.Observations[i].AskSize = nil
		s.Observations[i].Session = "UNKNOWN"
	}
	f, e := a.TryFill(id, s, at, cal, fee)
	if e != nil || f.Quantity != 10 || f.Price != "100.05" || f.ExecutionVersion != marketdata.ObservedPriceMode || a.Shares(asset) != 10 {
		t.Fatal(f, e, a)
	}
	if _, e = a.TryFill(id, s, at, cal, fee); e == nil {
		t.Fatal("filled order replayed")
	}
	id = uuid.NewString()
	sell := Action{AssetID: &asset, Side: ptr("SELL"), Quantity: ptr(int64(10)), Type: ptr("MARKET"), TIF: ptr("DAY")}
	if e = a.Submit(sell, id, at.Add(time.Second), cal, fee, false); e != nil {
		t.Fatal(e)
	}
	if _, e = a.TryFill(id, s, at.Add(2*time.Second), cal, fee); e == nil {
		t.Fatal("read before order accepted")
	}
	for i := range s.Observations {
		s.Observations[i].ObservedAt = at.Add(2 * time.Second)
	}
	f, e = a.TryFill(id, s, at.Add(2*time.Second), cal, fee)
	if e != nil || f.Quantity != 10 || f.Price != "99.95" || a.Shares(asset) != 0 || a.Unsettled == "0" {
		t.Fatal(f, e, a)
	}
	if e = a.Check(); e != nil {
		t.Fatal(e)
	}
}
