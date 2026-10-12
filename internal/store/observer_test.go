package store

import (
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/marketdata"
	"testing"
	"time"
)

func TestMissingValuationNeverBecomesZero(t *testing.T) {
	a := domain.NewAccount()
	a.Lots = []domain.Lot{{AssetID: "x", Quantity: 2, Cost: "100"}}
	v := valueWallet(Wallet{Account: a}, nil, nil, time.Now())
	if v.Complete || v.Equity != nil || v.PositionValue != nil {
		t.Fatal(v)
	}
	price := "60"
	asof := time.Now().Add(-2 * time.Minute)
	q := marketdata.Snapshot{Validation: marketdata.Validation{Verified: true, Median: &price, AsOf: asof}}
	v = valueWallet(Wallet{Account: a}, map[string]string{"x": "ABC"}, map[string]marketdata.Snapshot{"x": q}, time.Now())
	if !v.Complete || *v.PositionValue != "120" || *v.Equity != "100120" || v.Positions[0].Quality != "stale" {
		t.Fatal(v)
	}
}
