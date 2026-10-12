package actors

import (
	"github.com/caglarkc/trading/internal/domain"
	"testing"
)

func TestFixedPersonasAndVisibleAccount(t *testing.T) {
	p, e := LoadPersonas("../../config/personas.json")
	if e != nil || len(p.Items) != 8 {
		t.Fatal(e, p)
	}
	a := domain.NewAccount()
	a.QuoteUse["privateQuote"] = 1
	a.AppliedEvents["privateEvent"] = true
	v := visibleAccount(a)
	if v.QuoteUse != nil || v.AppliedEvents != nil || v.Available != "100000" {
		t.Fatal(v)
	}
	if a.QuoteUse["privateQuote"] != 1 {
		t.Fatal("source account mutated")
	}
}
