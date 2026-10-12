package domain

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type Account struct {
	DividendIncome string            `json:"dividend_income_usd"`
	Receivables    []Receivable      `json:"receivables"`
	Available      string            `json:"available_cash_usd"`
	Reserved       string            `json:"reserved_cash_usd"`
	Unsettled      string            `json:"unsettled_cash_usd"`
	Dividend       string            `json:"dividend_receivable_usd"`
	Realized       string            `json:"realized_pnl_usd"`
	Fees           string            `json:"fees_usd"`
	Orders         map[string]*Order `json:"orders"`
	Lots           []Lot             `json:"lots"`
	Settlements    []Settlement      `json:"settlements"`
	QuoteUse       map[string]int64  `json:"quote_use"`
	AppliedEvents  map[string]bool   `json:"applied_events"`
}
type Order struct {
	ID           string              `json:"id"`
	AssetID      string              `json:"asset_id"`
	Side         string              `json:"side"`
	Type         string              `json:"order_type"`
	TIF          string              `json:"time_in_force"`
	Quantity     int64               `json:"quantity"`
	Remaining    int64               `json:"remaining"`
	Filled       int64               `json:"filled"`
	Limit        *string             `json:"limit_price"`
	Stop         *string             `json:"stop_price"`
	Budget       *string             `json:"cash_budget_usd"`
	ReservedCash string              `json:"reserved_cash_usd"`
	State        string              `json:"state"`
	Reason       string              `json:"reason"`
	OCO          *string             `json:"oco_group_id"`
	Triggered    bool                `json:"triggered"`
	CreatedAt    time.Time           `json:"created_at"`
	ExpiresAt    time.Time           `json:"expires_at"`
	TriggeredAt  *time.Time          `json:"triggered_at"`
	DailyFees    map[string]DailyFee `json:"daily_fees"`
	Thesis       string              `json:"thesis"`
	EvidenceIDs  []string            `json:"evidence_ids"`
}
type DailyFee struct {
	Quantity   int64  `json:"quantity"`
	Notional   string `json:"notional"`
	Commission string `json:"commission"`
	Regulatory string `json:"regulatory"`
}
type Lot struct {
	ID         string    `json:"id"`
	AssetID    string    `json:"asset_id"`
	FillID     string    `json:"fill_id"`
	Quantity   int64     `json:"quantity"`
	Cost       string    `json:"cost_basis_usd"`
	AcquiredAt time.Time `json:"acquired_at"`
}
type Settlement struct {
	ID     string    `json:"id"`
	Amount string    `json:"amount_usd"`
	Due    time.Time `json:"due"`
	Done   bool      `json:"done"`
}

func NewAccount() Account {
	return Account{DividendIncome: "0", Receivables: []Receivable{}, Available: "100000", Reserved: "0", Unsettled: "0", Dividend: "0", Realized: "0", Fees: "0", Orders: map[string]*Order{}, Lots: []Lot{}, Settlements: []Settlement{}, QuoteUse: map[string]int64{}, AppliedEvents: map[string]bool{}}
}
func Money(s string) (decimal.Decimal, error) {
	d, e := decimal.NewFromString(s)
	if e != nil || d.Exponent() < -8 || len(s) > 40 {
		return decimal.Zero, fmt.Errorf("invalid_decimal")
	}
	return d, nil
}
func (a Account) Check() error {
	for _, s := range []string{a.Available, a.Reserved, a.Unsettled, a.Dividend, a.Fees} {
		d, e := Money(s)
		if e != nil || d.IsNegative() {
			return fmt.Errorf("negative_or_invalid_balance")
		}
	}
	reserve := decimal.Zero
	for _, o := range a.Orders {
		if o.Quantity <= 0 || o.Remaining < 0 || o.Filled < 0 || o.Filled+o.Remaining > o.Quantity {
			return fmt.Errorf("invalid_order_quantity")
		}
		d, e := Money(o.ReservedCash)
		if e != nil || d.IsNegative() {
			return fmt.Errorf("invalid_reserve")
		}
		reserve = reserve.Add(d)
	}
	r, _ := Money(a.Reserved)
	if !reserve.Equal(r) {
		return fmt.Errorf("reserve_mismatch")
	}
	for _, l := range a.Lots {
		c, e := Money(l.Cost)
		if l.Quantity < 0 || e != nil || c.IsNegative() || (l.Quantity == 0 && !c.IsZero()) {
			return fmt.Errorf("invalid_lot")
		}
	}
	for _, l := range a.Lots {
		if a.ReservedShares(l.AssetID, "") > a.Shares(l.AssetID) {
			return fmt.Errorf("share_reserve_mismatch")
		}
	}
	return nil
}
