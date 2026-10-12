package marketdata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ParserVersion = "public-dom-v1"

var numeric = regexp.MustCompile(`^[0-9][0-9,]*(\.[0-9]+)?$`)
var event = regexp.MustCompile(`(?i)(At close:|As of:|As of)\s*([^\n]+)`)

func validPrice(s string) *string {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	p, e := decimal.NewFromString(s)
	if e != nil || !p.IsPositive() {
		return nil
	}
	value := p.String()
	return &value
}

func HTMLPage(body string) Page {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return Page{}
	}
	p := Page{Method: "http"}
	var text strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" {
				return
			}
			attrs := map[string]string{}
			for _, a := range n.Attr {
				attrs[a.Key] = a.Val
			}
			if attrs["aria-hidden"] == "true" || attrs["hidden"] != "" || strings.Contains(attrs["style"], "display:none") {
				return
			}
			if n.Data == "title" {
				p.Title = nodeText(n)
			}
			if n.Data == "h1" {
				p.Heading = nodeText(n)
			}
			if p.RegularPrice == nil && strings.Contains(attrs["class"], "text-4xl") {
				p.RegularPrice = validPrice(nodeText(n))
			}
		}
		if n.Type == html.TextNode {
			value := strings.TrimSpace(n.Data)
			if value != "" {
				text.WriteString(value)
				text.WriteByte('\n')
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	p.Text = text.String()
	return p
}
func nodeText(n *html.Node) string {
	var b strings.Builder
	var w func(*html.Node)
	w = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			w(c)
		}
	}
	w(n)
	return strings.TrimSpace(b.String())
}

func Parse(source string, a Asset, url string, p Page, now time.Time) Observation {
	r := Observation{AssetID: a.ID, Symbol: a.Symbol, SourceID: source, SourceURL: url, Method: p.Method, ObservedAt: now.UTC(), PriceType: "UNKNOWN", Session: "UNKNOWN", ParserVersion: ParserVersion}
	alias := a.Symbol
	if source == "yahoo" && a.Symbol == "BRK.B" {
		alias = "BRK-B"
	}
	identity := regexp.MustCompile(`(?i)(^|[^A-Z0-9])` + regexp.QuoteMeta(alias) + `([^A-Z0-9]|$)`)
	r.IdentityVerified = identity.MatchString(p.Heading) || identity.MatchString(p.Title)
	t := p.Text
	r.Delayed = regexp.MustCompile(`(?i)\b(delayed quote|delayed price|quote delayed|delayed by|15.min.*delay)\b`).MatchString(t)
	r.Price = p.RegularPrice
	lines := strings.Split(t, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(p.Heading) && p.Heading != "" {
			scope := strings.Join(lines[max(0, i-8):min(len(lines), i+12)], "\n")
			if regexp.MustCompile(`(?i)\bUSD\b`).MatchString(scope) {
				r.Currency = "USD"
			}
			break
		}
	}
	if r.Price == nil {
		// Only the first quoted header near the asset heading is considered. A
		// day's high/low, NAV, prior close or related symbol is never a substitute.
		start := -1
		for i, l := range lines {
			if strings.TrimSpace(l) == strings.TrimSpace(p.Heading) && p.Heading != "" {
				start = i
				break
			}
		}
		if start >= 0 {
			for _, l := range lines[start+1 : min(len(lines), start+14)] {
				l = strings.TrimSpace(l)
				if numeric.MatchString(l) {
					r.Price = validPrice(l)
					break
				}
			}
		}
	}
	if source == "stockanalysis" || source == "yahoo" {
		matches := event.FindStringSubmatch(t)
		if len(matches) > 0 {
			kind := strings.ToLower(matches[1])
			raw := strings.TrimSpace(matches[2])
			r.EventTimeText = raw
			if strings.Contains(kind, "close") {
				r.Session = "REGULAR"
				r.PriceType = "CLOSE"
			} else if strings.Contains(strings.ToLower(t), "market open") || source == "stockanalysis" {
				r.Session = "REGULAR"
				r.PriceType = "LAST"
			}
			r.EventTime = parseEvent(raw, now, r.PriceType == "LAST" && strings.Contains(strings.ToLower(t), "market open"))
		}
	}
	// Webull's generic "Closed 19:30" clock is not proof that its headline is
	// the regular close. Preserve the number but leave unproved metadata blank.
	for _, field := range []string{"Bid", "Ask"} {
		rx := regexp.MustCompile(`(?i)\b` + field + `\s*\n?\s*([0-9][0-9,]*(?:\.[0-9]+)?)\s*(?:x\s*([0-9,]+))?`)
		m := rx.FindStringSubmatch(t)
		if len(m) == 0 {
			continue
		}
		v := validPrice(m[1])
		var size *int64
		if len(m) > 2 && m[2] != "" {
			n, e := strconv.ParseInt(strings.ReplaceAll(m[2], ",", ""), 10, 64)
			if e == nil && n > 0 {
				size = &n
			}
		}
		if field == "Bid" {
			r.Bid = v
			r.BidSize = size
		} else {
			r.Ask = v
			r.AskSize = size
		}
	}
	// No vendor-independent assumption about x100 being shares or about the
	// closing timestamp being the BBO timestamp. Missing evidence stays nil.
	if r.Price == nil {
		r.Error = "regular_price_not_visible"
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%v", a.ID, source, url, now.UTC().Format(time.RFC3339Nano), r.Price)))
	r.ID = hex.EncodeToString(hash[:])
	return r
}

func parseEvent(raw string, now time.Time, todayExplicit bool) *time.Time {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, ". Market Open."))
	raw = strings.ReplaceAll(raw, " at ", " ")
	location, _ := time.LoadLocation("America/New_York")
	for _, layout := range []string{"Jan 2, 2006, 3:04 PM MST", "Jan 2, 2006, 3:04:05 PM MST", "January 2, 2006 3:04:05 PM MST", "January 2, 2006 3:04 PM MST"} {
		if value, err := time.ParseInLocation(layout, raw, location); err == nil {
			v := value.UTC()
			return &v
		}
	}
	if todayExplicit {
		for _, layout := range []string{"3:04:05 PM MST", "3:04 PM MST"} {
			if value, err := time.ParseInLocation(layout, raw, location); err == nil {
				date := now.In(location)
				v := time.Date(date.Year(), date.Month(), date.Day(), value.Hour(), value.Minute(), value.Second(), 0, location).UTC()
				return &v
			}
		}
	}
	return nil
}
