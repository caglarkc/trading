package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func LoadFees(path string) (FeeProfile, error) {
	var f FeeProfile
	b, e := os.ReadFile(path)
	if e != nil {
		return f, e
	}
	if e = json.Unmarshal(b, &f); e != nil {
		return f, e
	}
	if f.Version == "" {
		return f, fmt.Errorf("fee_version_missing")
	}
	for _, v := range []string{f.PerShare, f.Minimum, f.Cap, f.SEC, f.CAT, f.TAF, f.Slippage, f.Tick} {
		d, e := Money(v)
		if e != nil || d.IsNegative() {
			return f, fmt.Errorf("invalid_fee_profile")
		}
	}
	if !dec(f.Tick).IsPositive() || dec(f.Slippage).GreaterThanOrEqual(dec("10000")) {
		return f, fmt.Errorf("invalid_execution_profile")
	}
	from, e := time.Parse("2006-01-02", f.From)
	if e != nil {
		return f, e
	}
	to, e := time.Parse("2006-01-02", f.To)
	if e != nil || to.Before(from) {
		return f, fmt.Errorf("invalid_fee_period")
	}
	return f, nil
}
