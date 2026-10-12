package marketdata

import (
	"github.com/shopspring/decimal"
	"sort"
	"time"
)

func Validate(asset string, rows []Observation, now time.Time, marketOpen bool) Validation {
	v := Validation{Reason: "price_unverified", AsOf: now, Observations: []string{}}
	if !marketOpen {
		v.Reason = "market_closed"
		return v
	}
	if len(rows) != 3 {
		v.Reason = "three_sources_required"
		return v
	}
	seen := map[string]bool{}
	prices := []decimal.Decimal{}
	var first, last time.Time
	for _, r := range rows {
		if seen[r.SourceID] || r.SourceID == "" {
			v.Reason = "duplicate_source"
			return v
		}
		seen[r.SourceID] = true
		if r.AssetID != asset || !r.IdentityVerified {
			v.Reason = "asset_identity_unverified"
			return v
		}
		if r.Currency != "" && r.Currency != "USD" || r.Error != "" {
			v.Reason = "quote_metadata_unverified"
			return v
		}
		if r.ObservedAt.IsZero() {
			v.Reason = "read_time_missing"
			return v
		}
		age := now.Sub(r.ObservedAt)
		if age < 0 || age > time.Minute {
			v.Reason = "quote_stale_or_future"
			return v
		}
		if first.IsZero() || r.ObservedAt.Before(first) {
			first = r.ObservedAt
		}
		if last.IsZero() || r.ObservedAt.After(last) {
			last = r.ObservedAt
		}
		if r.Price == nil {
			v.Reason = "price_missing"
			return v
		}
		p, err := decimal.NewFromString(*r.Price)
		if err != nil || !p.IsPositive() {
			v.Reason = "invalid_price"
			return v
		}
		prices = append(prices, p)
		v.Observations = append(v.Observations, r.ID)
	}
	if last.Sub(first) > time.Minute {
		v.Reason = "source_time_spread"
		return v
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LessThan(prices[j]) })
	spread := prices[2].Sub(prices[0]).Mul(decimal.NewFromInt(100)).Div(prices[1])
	med, pct := prices[1].String(), spread.String()
	v.Median = &med
	v.RangePercent = &pct
	if spread.GreaterThan(decimal.NewFromInt(1)) {
		v.Reason = "price_difference_exceeds_one_percent"
		return v
	}
	v.Verified = true
	v.Reason = "verified"
	return v
}

func VerifyLiquidity(r Observation, now time.Time) bool {
	if r.BBOTime == nil || r.Bid == nil || r.Ask == nil || r.BidSize == nil || r.AskSize == nil || r.SizeUnit != "shares" || r.VenueScope == "" || r.Currency != "USD" || r.Session != "REGULAR" || r.Delayed || !r.IdentityVerified {
		return false
	}
	age := now.Sub(*r.BBOTime)
	if age < 0 || age > time.Minute || *r.BidSize <= 0 || *r.AskSize <= 0 {
		return false
	}
	bid, e1 := decimal.NewFromString(*r.Bid)
	ask, e2 := decimal.NewFromString(*r.Ask)
	return e1 == nil && e2 == nil && bid.IsPositive() && ask.IsPositive() && !bid.GreaterThan(ask)
}
