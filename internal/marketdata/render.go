package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func (r *Reader) LoadHosts(path string) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, &r.Hosts)
}
func (r *Reader) permitted(target string) bool {
	u, e := url.Parse(target)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" && u.Port() != "443" {
		return false
	}
	for _, h := range r.Hosts {
		if allowed(strings.ToLower(u.Hostname()), h) {
			return true
		}
	}
	return false
}
func (r *Reader) remoteRender(ctx context.Context, target string) (Page, error) {
	var p Page
	b, _ := json.Marshal(map[string]string{"url": target})
	req, e := http.NewRequestWithContext(ctx, "POST", r.RenderURL+"/render", strings.NewReader(string(b)))
	if e != nil {
		return p, e
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 30 * time.Second}
	res, e := client.Do(req)
	if e != nil {
		return p, e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return p, fmt.Errorf("renderer_unavailable_%d", res.StatusCode)
	}
	e = json.NewDecoder(http.MaxBytesReader(nil, res.Body, 5*1024*1024)).Decode(&p)
	return p, e
}
func ServeRender(ctx context.Context, listen, chrome, hosts string) error {
	reader := NewReader(chrome)
	if e := reader.LoadHosts(hosts); e != nil {
		return e
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /render", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 2048)
		var in struct {
			URL string `json:"url"`
		}
		d := json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		if e := d.Decode(&in); e != nil || !reader.permitted(in.URL) {
			http.Error(w, "source_not_permitted", 400)
			return
		}
		if _, e := safeIPs(r.Context(), mustURLHost(in.URL)); e != nil {
			http.Error(w, "source_not_public", 400)
			return
		}
		p, e := reader.renderPage(r.Context(), in.URL)
		if e != nil {
			http.Error(w, e.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	})
	srv := &http.Server{Addr: listen, Handler: mux, ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 35 * time.Second}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(c)
	}()
	e := srv.ListenAndServe()
	if e == http.ErrServerClosed {
		return nil
	}
	return e
}
func mustURLHost(s string) string { u, _ := url.Parse(s); return u.Hostname() }

type Document struct {
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Text        string     `json:"text"`
	ObservedAt  time.Time  `json:"observed_at"`
	PublishedAt *time.Time `json:"published_at"`
	Method      string     `json:"method"`
	Truncated   bool       `json:"truncated"`
}

func (r *Reader) Document(ctx context.Context, target string) (Document, error) {
	p, e := r.page(ctx, target, false)
	return Document{URL: target, Title: p.Title, Text: p.Text, ObservedAt: time.Now().UTC(), PublishedAt: p.PublishedAt, Method: p.Method}, e
}
