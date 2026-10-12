package domain

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type CorporateEvent struct {
	ID             string    `json:"id"`
	AssetID        string    `json:"asset_id"`
	Kind           string    `json:"event_type"`
	EffectiveAt    time.Time `json:"effective_at"`
	Verified       bool      `json:"verified"`
	SourceURL      string    `json:"source_url"`
	Ratio          string    `json:"ratio"`
	PerShare       string    `json:"per_share"`
	PayAt          time.Time `json:"pay_at"`
	Classification string    `json:"classification"`
}
type Receivable struct {
	ID       string    `json:"id"`
	AssetID  string    `json:"asset_id"`
	Quantity int64     `json:"quantity"`
	Gross    string    `json:"gross_usd"`
	Tax      *string   `json:"tax_usd"`
	Net      *string   `json:"net_usd"`
	PayAt    time.Time `json:"pay_at"`
	Status   string    `json:"status"`
}

func (a *Account) Split(e CorporateEvent) error {
	if !e.Verified || e.Kind != "SPLIT" {
		return fmt.Errorf("corporate_action_unverified")
	}
	if a.AppliedEvents[e.ID] {
		return nil
	}
	ratio, err := Money(e.Ratio)
	if err != nil || !ratio.IsPositive() {
		return fmt.Errorf("invalid_split")
	}
	// Fractional ownership needs official cash-in-lieu terms; never guess cash.
	quantity := func(q int64) (int64, error) {
		n := decimal.NewFromInt(q).Mul(ratio)
		if !n.Equal(n.Truncate(0)) || n.GreaterThan(decimal.NewFromInt(9007199254740991)) {
			return 0, fmt.Errorf("fractional_split_unresolved")
		}
		return n.IntPart(), nil
	}
	for _, l := range a.Lots {
		if l.AssetID == e.AssetID {
			if _, err = quantity(l.Quantity); err != nil {
				return err
			}
		}
	}
	for _, o := range a.Orders {
		if o.AssetID == e.AssetID && Active(o) {
			if _, err = quantity(o.Quantity); err != nil {
				return err
			}
			if _, err = quantity(o.Remaining); err != nil {
				return err
			}
			if _, err = quantity(o.Filled); err != nil {
				return err
			}
		}
	}
	for i := range a.Lots {
		l := &a.Lots[i]
		if l.AssetID == e.AssetID {
			l.Quantity, _ = quantity(l.Quantity)
		}
	}
	for _, o := range a.Orders {
		if o.AssetID != e.AssetID || !Active(o) {
			continue
		}
		o.Quantity, _ = quantity(o.Quantity)
		o.Remaining, _ = quantity(o.Remaining)
		o.Filled, _ = quantity(o.Filled)
		if o.Limit != nil {
			p := dec(*o.Limit).Div(ratio).Mul(decimal.NewFromInt(100))
			if o.Side == "BUY" {
				p = p.Floor()
			} else {
				p = p.Ceil()
			}
			s := p.Div(decimal.NewFromInt(100)).String()
			o.Limit = &s
		}
		if o.Stop != nil {
			p := dec(*o.Stop).Div(ratio).Mul(decimal.NewFromInt(100))
			if o.Side == "BUY" {
				p = p.Ceil()
			} else {
				p = p.Floor()
			}
			s := p.Div(decimal.NewFromInt(100)).String()
			o.Stop = &s
		}
	}
	a.AppliedEvents[e.ID] = true
	return a.Check()
}
func (a *Account) DividendEntitlement(e CorporateEvent, quantity int64) error {
	if !e.Verified || e.Kind != "DIVIDEND" || e.PayAt.IsZero() || e.PayAt.Before(e.EffectiveAt) {
		return fmt.Errorf("dividend_unverified")
	}
	if a.AppliedEvents[e.ID] {
		return nil
	}
	p, err := Money(e.PerShare)
	if err != nil || !p.IsPositive() || quantity < 0 {
		return fmt.Errorf("invalid_dividend")
	}
	gross := p.Mul(decimal.NewFromInt(quantity)).Round(8)
	r := Receivable{ID: e.ID, AssetID: e.AssetID, Quantity: quantity, Gross: gross.String(), PayAt: e.PayAt, Status: "classification_unresolved"}
	income := gross
	if e.Classification == "ordinary_us" {
		tax := gross.Mul(decimal.RequireFromString("0.30")).Round(8).String()
		net := gross.Sub(dec(tax)).String()
		r.Tax = &tax
		r.Net = &net
		r.Status = "receivable"
		income = dec(net)
	}
	if a.DividendIncome == "" {
		a.DividendIncome = "0"
	}
	a.Dividend = dec(a.Dividend).Add(income).String()
	a.DividendIncome = dec(a.DividendIncome).Add(income).String()
	a.Receivables = append(a.Receivables, r)
	a.AppliedEvents[e.ID] = true
	return a.Check()
}
func (a *Account) PayDividends(now time.Time) bool {
	changed := false
	for i := range a.Receivables {
		r := &a.Receivables[i]
		if r.Status == "receivable" && r.Net != nil && !now.Before(r.PayAt) {
			n := dec(*r.Net)
			a.Dividend = dec(a.Dividend).Sub(n).String()
			a.Available = dec(a.Available).Add(n).String()
			r.Status = "paid"
			changed = true
		}
	}
	return changed
}
