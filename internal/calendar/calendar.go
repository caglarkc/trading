package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Calendar struct {
	Version            string   `json:"version"`
	MarketYears        []int    `json:"market_years"`
	SettlementYears    []int    `json:"settlement_years"`
	MarketHolidays     []string `json:"market_holidays"`
	EarlyClose         []string `json:"early_close"`
	SettlementHolidays []string `json:"settlement_holidays"`
}

var ET = func() *time.Location {
	l, e := time.LoadLocation("America/New_York")
	if e != nil {
		panic(e)
	}
	return l
}()

func Load(path string) (Calendar, error) {
	var c Calendar
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal(b, &c)
	return c, e
}
func contains[T comparable](a []T, x T) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}
func (c Calendar) Session(t time.Time) (time.Time, time.Time, bool, error) {
	d := t.In(ET)
	if !contains(c.MarketYears, d.Year()) {
		return time.Time{}, time.Time{}, false, fmt.Errorf("market_calendar_unverified")
	}
	if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday || contains(c.MarketHolidays, d.Format("2006-01-02")) {
		return time.Time{}, time.Time{}, false, nil
	}
	open := time.Date(d.Year(), d.Month(), d.Day(), 9, 30, 0, 0, ET)
	hour := 16
	if contains(c.EarlyClose, d.Format("2006-01-02")) {
		hour = 13
	}
	close := time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, ET)
	return open.UTC(), close.UTC(), true, nil
}
func (c Calendar) Open(t time.Time) bool {
	o, x, day, e := c.Session(t)
	return e == nil && day && !t.Before(o) && t.Before(x)
}
func (c Calendar) Settlement(t time.Time) (time.Time, error) {
	d := t.In(ET)
	for i := 0; i < 14; i++ {
		d = d.AddDate(0, 0, 1)
		if !contains(c.SettlementYears, d.Year()) {
			return time.Time{}, fmt.Errorf("settlement_calendar_unverified")
		}
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday && !contains(c.SettlementHolidays, d.Format("2006-01-02")) {
			return time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, ET).UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("settlement_date_unresolved")
}
func (c Calendar) Expiry(t time.Time, kind, tif string) (time.Time, error) {
	_, close, day, e := c.Session(t)
	if e != nil || !day {
		return time.Time{}, fmt.Errorf("market_closed")
	}
	if kind == "MARKET" {
		x := t.Add(120 * time.Second)
		if close.Before(x) {
			x = close
		}
		return x, nil
	}
	if tif == "DAY" {
		return close, nil
	}
	d := t.In(ET).AddDate(0, 0, 90)
	for i := 0; i < 8; i++ {
		_, x, day, e := c.Session(d)
		if e != nil {
			return time.Time{}, e
		}
		if day {
			return x, nil
		}
		d = d.AddDate(0, 0, 1)
	}
	return time.Time{}, fmt.Errorf("expiry_unresolved")
}
