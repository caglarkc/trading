package marketdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
)

type Reader struct {
	Client     *http.Client
	ChromePath string
	mu         sync.Mutex
	gates      map[string]*sync.Mutex
	last       map[string]time.Time
	render     chan struct{}
	failures   map[string]int
	cooldown   map[string]time.Time
	slots      chan struct{}
}

func NewReader(chromePath string) *Reader {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(address)
		if e != nil {
			return nil, e
		}
		ips, e := safeIPs(ctx, host)
		if e != nil {
			return nil, e
		}
		var last error
		for _, ip := range ips {
			c, e := (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if e == nil {
				return c, nil
			}
			last = e
		}
		return nil, last
	}
	r := &Reader{ChromePath: chromePath, gates: map[string]*sync.Mutex{}, last: map[string]time.Time{}, render: make(chan struct{}, 1), failures: map[string]int{}, cooldown: map[string]time.Time{}, slots: make(chan struct{}, 8)}
	r.Client = &http.Client{Transport: transport, Timeout: 12 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("redirect limit")
		}
		if req.URL.Scheme != "https" || !allowed(req.URL.Hostname(), via[0].URL.Hostname()) {
			return fmt.Errorf("redirect outside permitted source")
		}
		_, e := safeIPs(req.Context(), req.URL.Hostname())
		return e
	}}
	return r
}
func allowed(host, original string) bool {
	return host == original || strings.HasSuffix(host, "."+strings.TrimPrefix(original, "www."))
}
func safeIPs(ctx context.Context, host string) ([]net.IP, error) {
	ips, e := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if e != nil {
		return nil, e
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no public address")
	}
	for _, ip := range ips {
		if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return nil, fmt.Errorf("private or non-public source address")
		}
	}
	return ips, nil
}
func (r *Reader) gate(host string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.gates[host] == nil {
		r.gates[host] = &sync.Mutex{}
	}
	return r.gates[host]
}
func (r *Reader) page(ctx context.Context, target string, quote bool) (page Page, err error) {
	u, e := url.Parse(target)
	if e != nil || u.Scheme != "https" || u.User != nil {
		return Page{}, fmt.Errorf("invalid public URL")
	}
	if _, e = safeIPs(ctx, u.Hostname()); e != nil {
		return Page{}, e
	}
	g := r.gate(u.Hostname())
	g.Lock()
	defer g.Unlock()
	r.mu.Lock()
	until := r.cooldown[u.Hostname()]
	r.mu.Unlock()
	if until.After(time.Now()) {
		return Page{}, fmt.Errorf("source_cooldown_until %s", until.UTC().Format(time.RFC3339))
	}
	defer func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if err != nil {
			r.failures[u.Hostname()]++
			if r.failures[u.Hostname()] >= 3 {
				r.cooldown[u.Hostname()] = time.Now().Add(5 * time.Minute)
			}
		} else {
			r.failures[u.Hostname()] = 0
		}
	}()
	r.mu.Lock()
	wait := time.Until(r.last[u.Hostname()].Add(time.Second))
	r.mu.Unlock()
	if wait > 0 {
		select {
		case <-ctx.Done():
			return Page{}, ctx.Err()
		case <-time.After(wait):
		}
	}
	r.mu.Lock()
	r.last[u.Hostname()] = time.Now()
	r.mu.Unlock()
	req, e := http.NewRequestWithContext(ctx, "GET", target, nil)
	if e != nil {
		return Page{}, e
	}
	req.Header.Set("User-Agent", "TradingResearch/0.1 (public source verification)")
	select {
	case r.slots <- struct{}{}:
	case <-ctx.Done():
		return Page{}, ctx.Err()
	}
	response, e := r.Client.Do(req)
	<-r.slots
	if e == nil {
		defer response.Body.Close()
		if response.StatusCode == http.StatusTooManyRequests {
			delay := 5 * time.Minute
			if value, e := time.ParseDuration(response.Header.Get("Retry-After") + "s"); e == nil && value > delay {
				delay = value
			}
			if date, e := http.ParseTime(response.Header.Get("Retry-After")); e == nil && time.Until(date) > delay {
				delay = time.Until(date)
			}
			r.mu.Lock()
			r.cooldown[u.Hostname()] = time.Now().Add(delay)
			r.mu.Unlock()
			return Page{}, fmt.Errorf("source_429 retry_after=%s", response.Header.Get("Retry-After"))
		}
		if response.StatusCode == 401 || response.StatusCode == 403 || response.StatusCode == 404 {
			return Page{}, fmt.Errorf("source_http_%d", response.StatusCode)
		}
		if response.StatusCode == http.StatusOK {
			b, err := io.ReadAll(io.LimitReader(response.Body, 5*1024*1024+1))
			if err != nil {
				return Page{}, err
			}
			if len(b) > 5*1024*1024 {
				return Page{}, fmt.Errorf("body exceeds 5 MiB")
			}
			p := HTMLPage(string(b))
			if p.RegularPrice != nil || (!quote && len(p.Text) > 100) {
				return p, nil
			}
		}
	}
	if r.ChromePath == "" {
		return Page{}, fmt.Errorf("source inaccessible or needs render")
	}
	// Normal rendering is attempted only where permitted, never to solve a
	// CAPTCHA, login wall, certificate warning or explicit HTTP 429 restriction.
	return r.renderPage(ctx, target)
}
func (r *Reader) renderPage(ctx context.Context, target string) (Page, error) {
	select {
	case r.render <- struct{}{}:
	case <-ctx.Done():
		return Page{}, ctx.Err()
	}
	defer func() { <-r.render }()
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.ExecPath(r.ChromePath), chromedp.NoSandbox, chromedp.Flag("disable-dev-shm-usage", true))
	alloc, stop := chromedp.NewExecAllocator(ctx, opts...)
	defer stop()
	browser, closeBrowser := chromedp.NewContext(alloc)
	defer closeBrowser()
	events := chromedp.Events(browser, fetch.RequestPaused)
	go func() {
		for paused, eventErr := range events {
			if eventErr != nil {
				return
			}
			go func(paused fetch.EventRequestPaused) {
				u, e := url.Parse(paused.Request.URL)
				if e == nil && (u.Scheme == "https" || u.Scheme == "http") {
					_, e = safeIPs(browser, u.Hostname())
				} else {
					e = fmt.Errorf("non-public resource")
				}
				if e != nil {
					_, _ = chromedp.Call(browser, fetch.FailRequest, fetch.FailRequestParams{RequestID: paused.RequestID, ErrorReason: "BlockedByClient"})
				} else {
					_, _ = chromedp.Call(browser, fetch.ContinueRequest, fetch.ContinueRequestParams{RequestID: paused.RequestID})
				}
			}(paused)
		}
	}()
	if _, err := chromedp.Call(browser, fetch.Enable, fetch.EnableParams{}); err != nil {
		return Page{}, err
	}
	type renderedPage struct {
		Text, Heading, Title string
		Price                *string
	}
	u, _ := url.Parse(target)
	symbol := strings.ToUpper(path.Base(strings.TrimSuffix(u.Path, "/")))
	if strings.Contains(u.Hostname(), "webull") {
		_, symbol, _ = strings.Cut(symbol, "-")
	}
	quoted, _ := json.Marshal(symbol)
	js := fmt.Sprintf(`(()=>{const s=%s;const h=[...document.querySelectorAll('h1')];const heading=h.find(n=>n.innerText.includes('('+s+')'))||h.at(-1);const p=document.querySelector('fin-streamer[data-symbol="'+s+'"][data-field="regularMarketPrice"], [class~="text-4xl"]');return {Text:document.body.innerText,Heading:heading?.innerText||'',Title:document.title,Price:p?.innerText||null}})()`, quoted)
	err := chromedp.Do(browser, chromedp.Navigate(target), chromedp.WaitReady(chromedp.CSS("body")))
	if err != nil {
		return Page{}, err
	}
	if _, err = chromedp.Run(browser, chromedp.Poll[bool](`document.body.innerText.length>300`)); err != nil {
		return Page{}, err
	}
	value, err := chromedp.Run(browser, chromedp.Evaluate[renderedPage](js))
	if err != nil {
		return Page{}, err
	}
	if regexpCaptcha(value.Title + "\n" + value.Text) {
		return Page{}, fmt.Errorf("source_access_challenge")
	}
	price := value.Price
	if price != nil {
		price = validPrice(*price)
	}
	return Page{Text: value.Text, Heading: value.Heading, Title: value.Title, RegularPrice: price, Method: "chromium_visible_dom"}, nil
}
func (r *Reader) Read(ctx context.Context, a Asset, source string) Observation {
	target := a.Aliases[source]
	now := time.Now().UTC()
	failure := Observation{AssetID: a.ID, Symbol: a.Symbol, SourceID: source, SourceURL: target, ObservedAt: now, PriceType: "UNKNOWN", Session: "UNKNOWN", ParserVersion: ParserVersion}
	if target == "" {
		failure.Error = "source_alias_missing"
		return failure
	}
	p, e := r.page(ctx, target, true)
	if e != nil {
		failure.Error = e.Error()
		h := sha256.Sum256([]byte(a.ID + source + now.String()))
		failure.ID = hex.EncodeToString(h[:])
		return failure
	}
	return Parse(source, a, target, p, time.Now().UTC())
}
func (r *Reader) Collect(ctx context.Context, a Asset, marketOpen bool) Snapshot {
	sources := []string{"stockanalysis", "yahoo", "webull"}
	rows := make([]Observation, 3)
	var wg sync.WaitGroup
	for i, source := range sources {
		wg.Add(1)
		go func(i int, source string) { defer wg.Done(); rows[i] = r.Read(ctx, a, source) }(i, source)
	}
	wg.Wait()
	now := time.Now().UTC()
	v := Validate(a.ID, rows, now, marketOpen)
	liq := false
	for _, row := range rows {
		liq = liq || VerifyLiquidity(row, now)
	}
	return Snapshot{a, rows, v, liq}
}
func (r *Reader) PublicPage(ctx context.Context, target string) (string, error) {
	p, e := r.page(ctx, target, false)
	return p.Text, e
}

func regexpCaptcha(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "verify you are human") || strings.Contains(text, "complete the captcha") || strings.Contains(text, "checking your browser before accessing")
}
func DecodeAssets(data []byte) ([]Asset, error) {
	var a []Asset
	e := json.Unmarshal(data, &a)
	return a, e
}
