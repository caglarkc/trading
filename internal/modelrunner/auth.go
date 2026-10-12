package modelrunner

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Auth is owned by the trusted runner. Only active CLI invocations receive a
// private snapshot; it is reconciled and removed when that invocation ends.
type Auth struct {
	Path       string
	mu         sync.Mutex
	active     int
	poisoned   bool
	refreshing bool
}

type authLease struct {
	owner   *Auth
	path    string
	initial [32]byte
}

func authFresh(b []byte, now time.Time) bool {
	var x struct {
		Tokens struct {
			Access string `json:"access_token"`
		} `json:"tokens"`
		LastRefresh time.Time `json:"last_refresh"`
	}
	if json.Unmarshal(b, &x) != nil {
		return false
	}
	pieces := strings.Split(x.Tokens.Access, ".")
	if len(pieces) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(pieces[1])
	if err != nil {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return false
	}
	// Unknown/near-expiry credentials run singly, allowing the official CLI to
	// refresh without another invocation racing the refresh-token rotation.
	return time.Unix(claims.Exp, 0).After(now.Add(30*time.Minute)) && !x.LastRefresh.IsZero() && now.Sub(x.LastRefresh) < 5*24*time.Hour
}

func (a *Auth) acquire(ctx context.Context, home string) (*authLease, error) {
	for {
		a.mu.Lock()
		if a.refreshing {
			a.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		if a.poisoned {
			a.mu.Unlock()
			return nil, &Error{"auth_required", "credential refresh conflict; trusted reauthentication required"}
		}
		b, err := os.ReadFile(a.Path)
		if err != nil {
			a.mu.Unlock()
			return nil, &Error{"auth_required", "credential cache unavailable"}
		}
		max := 1
		if authFresh(b, time.Now()) {
			max = 2
		}
		if a.active < max {
			p := filepath.Join(home, "auth.json")
			err = atomicWrite(p, b, 0600)
			if err == nil {
				a.active++
			}
			a.mu.Unlock()
			if err != nil {
				return nil, err
			}
			return &authLease{a, p, sha256.Sum256(b)}, nil
		}
		a.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (l *authLease) release() error {
	a := l.owner
	a.mu.Lock()
	defer a.mu.Unlock()
	defer func() { a.active-- }()
	defer os.Remove(l.path)
	b, err := os.ReadFile(l.path)
	if err != nil {
		return err
	}
	changed := sha256.Sum256(b)
	if changed == l.initial {
		return nil
	}
	canonical, err := os.ReadFile(a.Path)
	if err != nil {
		return err
	}
	current := sha256.Sum256(canonical)
	if current != l.initial && current != changed {
		a.poisoned = true
		return &Error{"auth_required", "concurrent credential rotation conflict"}
	}
	if current == changed {
		return nil
	}
	return atomicWrite(a.Path, b, 0600)
}

func atomicWrite(path string, b []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".atomic-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	err = errors.Join(f.Chmod(mode), writeAll(f, b), f.Sync(), f.Close())
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}
func writeAll(f *os.File, b []byte) error { _, e := f.Write(b); return e }
