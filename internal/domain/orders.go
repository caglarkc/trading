package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Action struct {
	Action         string   `json:"action"`
	ClientID       string   `json:"client_action_id"`
	AssetID        *string  `json:"asset_id"`
	Side           *string  `json:"side"`
	Quantity       *int64   `json:"quantity"`
	Type           *string  `json:"order_type"`
	TIF            *string  `json:"time_in_force"`
	Target         *string  `json:"target_order_id"`
	OCO            *string  `json:"oco_group_id"`
	Limit          *string  `json:"limit_price"`
	Stop           *string  `json:"stop_price"`
	Budget         *string  `json:"cash_budget_usd"`
	Thesis         string   `json:"thesis"`
	EvidenceIDs    []string `json:"evidence_ids"`
	EvidenceStatus string   `json:"evidence_status"`
	Catalyst       string   `json:"expected_catalyst"`
	Exit           string   `json:"exit_condition"`
	Hold           *string  `json:"hold_reason"`
	Depends        []string `json:"depends_on"`
}
type Decision struct {
	Actions  []Action            `json:"actions"`
	Memory   string              `json:"memory_update"`
	Learning LearningApplication `json:"learning_application"`
}
type LearningApplication struct {
	Version     int64    `json:"strategy_version"`
	RuleIDs     []string `json:"applied_rule_ids"`
	Explanation string   `json:"explanation"`
}
type Receipt struct {
	ActionID string    `json:"client_action_id,omitempty"`
	Order    *Order    `json:"order,omitempty"`
	Status   string    `json:"status"`
	Error    string    `json:"error,omitempty"`
	Fill     *Fill     `json:"fill,omitempty"`
	Account  *Account  `json:"account,omitempty"`
	AsOf     time.Time `json:"as_of"`
}
type Fill struct {
	ID               string    `json:"id"`
	OrderID          string    `json:"order_id"`
	AssetID          string    `json:"asset_id"`
	Quantity         int64     `json:"quantity"`
	Price            string    `json:"price"`
	Commission       string    `json:"commission_usd"`
	Regulatory       string    `json:"regulatory_usd"`
	OccurredAt       time.Time `json:"occurred_at"`
	QuoteID          string    `json:"quote_id"`
	ValidationIDs    []string  `json:"validation_ids"`
	FeeVersion       string    `json:"fee_version"`
	ExecutionVersion string    `json:"execution_version"`
}
type FeeProfile struct {
	Version  string `json:"version"`
	From     string `json:"effective_from"`
	To       string `json:"effective_to"`
	PerShare string `json:"per_share"`
	Minimum  string `json:"minimum"`
	Cap      string `json:"cap_rate"`
	SEC      string `json:"sec_sell_rate"`
	CAT      string `json:"cat_share_rate"`
	TAF      string `json:"taf_share_rate"`
	Slippage string `json:"slippage_bps"`
	Tick     string `json:"price_tick"`
}

func dec(s string) decimal.Decimal {
	d, e := Money(s)
	if e != nil {
		panic(e)
	}
	return d
}
func (f FeeProfile) Valid(t time.Time) bool {
	d := t.In(calendar.ET).Format("2006-01-02")
	return f.Version != "" && d >= f.From && d <= f.To
}
func (f FeeProfile) Fees(q int64, notional decimal.Decimal, side string) (decimal.Decimal, decimal.Decimal) {
	shares := decimal.NewFromInt(q)
	c := decimal.Min(decimal.Max(shares.Mul(dec(f.PerShare)), dec(f.Minimum)), notional.Mul(dec(f.Cap))).RoundCeil(2)
	reg := shares.Mul(dec(f.CAT)).RoundCeil(2)
	if side == "SELL" {
		reg = reg.Add(notional.Mul(dec(f.SEC)).RoundCeil(2)).Add(shares.Mul(dec(f.TAF)).RoundCeil(2))
	}
	return c, reg
}
func Active(o *Order) bool {
	return o.State == "accepted" || o.State == "waiting" || o.State == "partially_filled"
}
func (a *Account) Shares(asset string) int64 {
	var n int64
	for _, l := range a.Lots {
		if l.AssetID == asset {
			n += l.Quantity
		}
	}
	return n
}
func (a *Account) ReservedShares(asset, excludeGroup string) int64 {
	groups := map[string]int64{}
	var n int64
	for _, o := range a.Orders {
		if !Active(o) || o.Side != "SELL" || o.AssetID != asset {
			continue
		}
		if o.OCO == nil {
			n += o.Remaining
		} else if *o.OCO != excludeGroup {
			if o.Remaining > groups[*o.OCO] {
				groups[*o.OCO] = o.Remaining
			}
		}
	}
	for _, q := range groups {
		n += q
	}
	return n
}
func (a *Account) Cancel(id string) error {
	o := a.Orders[id]
	if o == nil || !Active(o) {
		return fmt.Errorf("order_not_open")
	}
	r := dec(o.ReservedCash)
	a.Available = dec(a.Available).Add(r).String()
	a.Reserved = dec(a.Reserved).Sub(r).String()
	o.ReservedCash = "0"
	o.State = "cancelled"
	return nil
}
func (a *Account) Submit(x Action, id string, now time.Time, cal calendar.Calendar, fee FeeProfile, allowed bool) error {
	if x.AssetID == nil || x.Side == nil || x.Quantity == nil || x.Type == nil || x.TIF == nil || *x.Quantity <= 0 || *x.Quantity > 9007199254740991 {
		return fmt.Errorf("invalid_order")
	}
	if _, e := uuid.Parse(*x.AssetID); e != nil {
		return fmt.Errorf("invalid_asset")
	}
	if *x.Side != "BUY" && *x.Side != "SELL" {
		return fmt.Errorf("invalid_side")
	}
	if *x.TIF != "DAY" && *x.TIF != "GTC" {
		return fmt.Errorf("invalid_tif")
	}
	if *x.Type != "MARKET" && *x.Type != "LIMIT" && *x.Type != "STOP_MARKET" && *x.Type != "STOP_LIMIT" {
		return fmt.Errorf("invalid_order_type")
	}
	if *x.Type == "MARKET" && *x.TIF != "DAY" {
		return fmt.Errorf("market_tif_invalid")
	}
	if *x.Side == "BUY" && !allowed {
		return fmt.Errorf("asset_outside_universe")
	}
	limitRequired := *x.Type == "LIMIT" || *x.Type == "STOP_LIMIT"
	stopRequired := *x.Type == "STOP_MARKET" || *x.Type == "STOP_LIMIT"
	for _, p := range []struct {
		required bool
		v        *string
	}{{limitRequired, x.Limit}, {stopRequired, x.Stop}} {
		if p.required != (p.v != nil) {
			return fmt.Errorf("invalid_price_fields")
		}
		if p.v != nil {
			d, e := Money(*p.v)
			if e != nil || !d.IsPositive() || !d.Mod(dec(fee.Tick)).IsZero() {
				return fmt.Errorf("invalid_price_increment")
			}
		}
	}
	expiry, e := cal.Expiry(now, *x.Type, *x.TIF)
	if e != nil {
		return e
	}
	reserve := "0"
	if *x.Side == "BUY" {
		if x.OCO != nil || x.Budget == nil {
			return fmt.Errorf("buy_budget_required")
		}
		b, e := Money(*x.Budget)
		if e != nil || !b.IsPositive() || b.GreaterThan(dec(a.Available)) {
			return fmt.Errorf("insufficient_settled_cash")
		}
		if limitRequired {
			if !fee.Valid(now) {
				return fmt.Errorf("fee_profile_unverified")
			}
			n := dec(*x.Limit).Mul(decimal.NewFromInt(*x.Quantity))
			c, r := fee.Fees(*x.Quantity, n, "BUY")
			if b.LessThan(n.Add(c).Add(r)) {
				return fmt.Errorf("budget_below_limit_cost")
			}
		}
		reserve = b.String()
		a.Available = dec(a.Available).Sub(b).String()
		a.Reserved = dec(a.Reserved).Add(b).String()
	} else {
		if x.Budget != nil {
			return fmt.Errorf("sell_budget_invalid")
		}
		group := ""
		if x.OCO != nil {
			if _, e := uuid.Parse(*x.OCO); e != nil {
				return fmt.Errorf("invalid_oco")
			}
			group = *x.OCO
			var peers []*Order
			for _, o := range a.Orders {
				if o.OCO != nil && *o.OCO == group && Active(o) {
					peers = append(peers, o)
				}
			}
			if len(peers) > 1 {
				return fmt.Errorf("oco_has_two_legs")
			}
			for _, o := range peers {
				if o.AssetID != *x.AssetID || o.Side != "SELL" || o.Remaining != *x.Quantity || !(o.Type == "LIMIT" && stopRequired || *x.Type == "LIMIT" && (o.Type == "STOP_MARKET" || o.Type == "STOP_LIMIT")) {
					return fmt.Errorf("oco_leg_mismatch")
				}
			}
		}
		if *x.Quantity > a.Shares(*x.AssetID)-a.ReservedShares(*x.AssetID, group) {
			return fmt.Errorf("insufficient_shares")
		}
	}
	a.Orders[id] = &Order{ID: id, AssetID: *x.AssetID, Side: *x.Side, Type: *x.Type, TIF: *x.TIF, Quantity: *x.Quantity, Remaining: *x.Quantity, Limit: x.Limit, Stop: x.Stop, Budget: x.Budget, ReservedCash: reserve, State: "accepted", OCO: x.OCO, CreatedAt: now, ExpiresAt: expiry, DailyFees: map[string]DailyFee{}, Thesis: x.Thesis, EvidenceIDs: x.EvidenceIDs}
	return a.Check()
}
func (a *Account) Expire(now time.Time) {
	for _, o := range a.Orders {
		if Active(o) && !now.Before(o.ExpiresAt) {
			a.Cancel(o.ID)
			o.State = "expired"
		}
	}
}
func QuoteKey(o marketdata.Observation) string {
	t := ""
	if o.BBOTime != nil {
		t = o.BBOTime.UTC().Format(time.RFC3339Nano)
	}
	b, _ := json.Marshal([]any{o.AssetID, o.SourceID, o.VenueScope, t, o.Bid, o.Ask, o.BidSize, o.AskSize, o.SizeUnit})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (a *Account) TryFill(id string, s marketdata.Snapshot, now time.Time, cal calendar.Calendar, fee FeeProfile) (*Fill, error) {
	a.Expire(now)
	o := a.Orders[id]
	if o == nil || !Active(o) {
		return nil, fmt.Errorf("order_not_open")
	}
	if !cal.Open(now) {
		return nil, fmt.Errorf("market_closed")
	}
	if s.Asset.ID != o.AssetID {
		return nil, fmt.Errorf("asset_mismatch")
	}
	v := marketdata.Validate(o.AssetID, s.Observations, now, true)
	if !v.Verified {
		return nil, fmt.Errorf("price_unverified")
	}
	if !fee.Valid(now) {
		return nil, fmt.Errorf("fee_profile_unverified")
	}
	for _, row := range s.Observations {
		if row.ObservedAt.Before(o.CreatedAt) {
			return nil, fmt.Errorf("quote_predates_order")
		}
	}
	if o.Stop != nil && !o.Triggered {
		last := dec(*v.Median)
		stop := dec(*o.Stop)
		if o.Side == "BUY" && last.LessThan(stop) || o.Side == "SELL" && last.GreaterThan(stop) {
			return nil, fmt.Errorf("stop_not_triggered")
		}
		o.Triggered = true
		t := now
		o.TriggeredAt = &t
		if o.Type == "STOP_MARKET" {
			_, close, _, _ := cal.Session(now)
			o.ExpiresAt = now.Add(120 * time.Second)
			if close.Before(o.ExpiresAt) {
				o.ExpiresAt = close
			}
		}
	}
	var quote *marketdata.Observation
	key := ""
	size := o.Remaining
	raw := dec(*v.Median)
	volumeKey := ""
	executionVersion := marketdata.ObservedPriceMode
	if s.ExecutionMode == marketdata.ObservedPriceMode {
		b, _ := json.Marshal([]any{s.Asset.ID, v.Observations, v.Median})
		h := sha256.Sum256(b)
		key = hex.EncodeToString(h[:])
	} else {
		// Retain the historic BBO execution model for existing fixtures/replays.
		for i := range s.Observations {
			row := &s.Observations[i]
			if marketdata.VerifyLiquidity(*row, now) && row.BBOTime != nil && !row.BBOTime.Before(o.CreatedAt) {
				quote = row
				break
			}
		}
		if quote == nil {
			return nil, fmt.Errorf("liquidity_unverified")
		}
		key = QuoteKey(*quote)
		size = *quote.AskSize
		raw = dec(*quote.Ask)
		if o.Side == "SELL" {
			size = *quote.BidSize
			raw = dec(*quote.Bid)
		}
		executionVersion = "bbo-adverse-5bps-v1"
	}
	sideKey := key + ":" + o.Side
	factor := decimal.NewFromInt(1).Add(dec(fee.Slippage).Div(decimal.NewFromInt(10000)))
	if o.Side == "SELL" {
		factor = decimal.NewFromInt(1).Sub(dec(fee.Slippage).Div(decimal.NewFromInt(10000)))
	}
	price := raw.Mul(factor).Div(dec(fee.Tick))
	if o.Side == "BUY" {
		price = price.Ceil()
	} else {
		price = price.Floor()
	}
	price = price.Mul(dec(fee.Tick))
	if !price.IsPositive() {
		return nil, fmt.Errorf("invalid_execution_price")
	}
	if o.Limit != nil && (o.Side == "BUY" && price.GreaterThan(dec(*o.Limit)) || o.Side == "SELL" && price.LessThan(dec(*o.Limit))) {
		return nil, fmt.Errorf("limit_not_reached")
	}
	qty := o.Remaining
	if quote != nil {
		qty = min(o.Remaining, size-a.QuoteUse[sideKey])
		if quote.MinuteVolume != nil && quote.VolumeMinute != nil && quote.VolumeScope == quote.VenueScope && *quote.MinuteVolume > 0 && now.Sub(*quote.VolumeMinute) >= 0 && now.Sub(*quote.VolumeMinute) <= 60*time.Second {
			volumeKey = "volume:" + o.AssetID + ":" + quote.SourceID + ":" + quote.VolumeMinute.UTC().Format(time.RFC3339)
			qty = min(qty, *quote.MinuteVolume/100-a.QuoteUse[volumeKey])
		}
	}
	if qty <= 0 {
		return nil, fmt.Errorf("liquidity_consumed")
	}
	day := now.In(calendar.ET).Format("2006-01-02")
	prev := o.DailyFees[day]
	pn := decimal.Zero
	pc := decimal.Zero
	pr := decimal.Zero
	if prev.Notional != "" {
		pn = dec(prev.Notional)
		pc = dec(prev.Commission)
		pr = dec(prev.Regulatory)
	}
	calc := func(q int64) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
		n := price.Mul(decimal.NewFromInt(q))
		c, r := fee.Fees(prev.Quantity+q, pn.Add(n), o.Side)
		return n, c.Sub(pc), r.Sub(pr)
	}
	if o.Side == "BUY" {
		budget := dec(o.ReservedCash)
		lo, hi := int64(0), qty
		for lo < hi {
			m := lo + (hi-lo+1)/2
			n, c, r := calc(m)
			if n.Add(c).Add(r).LessThanOrEqual(budget) {
				lo = m
			} else {
				hi = m - 1
			}
		}
		qty = lo
		if qty == 0 {
			return nil, fmt.Errorf("insufficient_reserved_cash")
		}
	}
	due, e := cal.Settlement(now)
	if e != nil {
		return nil, e
	}
	n, c, r := calc(qty)
	fees := c.Add(r)
	fill := &Fill{uuid.NewString(), id, o.AssetID, qty, price.String(), c.String(), r.String(), now, key, v.Observations, fee.Version, executionVersion}
	if o.Side == "BUY" {
		cost := n.Add(fees)
		o.ReservedCash = dec(o.ReservedCash).Sub(cost).String()
		a.Reserved = dec(a.Reserved).Sub(cost).String()
		a.Lots = append(a.Lots, Lot{uuid.NewString(), o.AssetID, fill.ID, qty, cost.String(), now})
	} else {
		remaining := qty
		basis := decimal.Zero
		for i := range a.Lots {
			l := &a.Lots[i]
			if l.AssetID != o.AssetID || l.Quantity == 0 {
				continue
			}
			q := min(remaining, l.Quantity)
			cost := dec(l.Cost)
			part := cost
			if q < l.Quantity {
				part = cost.Mul(decimal.NewFromInt(q)).Div(decimal.NewFromInt(l.Quantity)).Round(8)
			}
			l.Quantity -= q
			l.Cost = cost.Sub(part).String()
			basis = basis.Add(part)
			remaining -= q
			if remaining == 0 {
				break
			}
		}
		if remaining != 0 {
			return nil, fmt.Errorf("position_invariant")
		}
		net := n.Sub(fees)
		a.Unsettled = dec(a.Unsettled).Add(net).String()
		a.Realized = dec(a.Realized).Add(net.Sub(basis)).String()
		a.Settlements = append(a.Settlements, Settlement{fill.ID, net.String(), due, false})
	}
	a.Fees = dec(a.Fees).Add(fees).String()
	a.QuoteUse[sideKey] += qty
	if volumeKey != "" {
		a.QuoteUse[volumeKey] += qty
	}
	o.Filled += qty
	o.Remaining -= qty
	o.DailyFees[day] = DailyFee{prev.Quantity + qty, pn.Add(n).String(), pc.Add(c).String(), pr.Add(r).String()}
	o.State = "partially_filled"
	if o.OCO != nil {
		for _, p := range a.Orders {
			if p.ID != o.ID && Active(p) && p.OCO != nil && *p.OCO == *o.OCO {
				p.Remaining -= qty
				if p.Remaining == 0 {
					p.State = "cancelled"
				}
			}
		}
	}
	if o.Remaining == 0 {
		o.State = "filled"
		unused := dec(o.ReservedCash)
		o.ReservedCash = "0"
		a.Reserved = dec(a.Reserved).Sub(unused).String()
		a.Available = dec(a.Available).Add(unused).String()
	}
	if e = a.Check(); e != nil {
		return nil, e
	}
	return fill, nil
}
func (a *Account) Settle(now time.Time) bool {
	changed := false
	for i := range a.Settlements {
		s := &a.Settlements[i]
		if !s.Done && !now.Before(s.Due) {
			d := dec(s.Amount)
			a.Available = dec(a.Available).Add(d).String()
			a.Unsettled = dec(a.Unsettled).Sub(d).String()
			s.Done = true
			changed = true
		}
	}
	return changed
}
