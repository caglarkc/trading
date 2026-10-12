package modelrunner

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// The official app-server is used only for managed account refresh, never as
// a second model/order path. Refresh runs exclusively with no active CLI calls.
func (a *Auth) EnsureFresh(ctx context.Context, root string, l Launcher) (err error) {
	for {
		a.mu.Lock()
		if a.poisoned {
			a.mu.Unlock()
			return &Error{"auth_required", "credential coordinator requires reauthentication"}
		}
		b, e := os.ReadFile(a.Path)
		if e != nil {
			a.mu.Unlock()
			return &Error{"auth_required", "login cache unavailable"}
		}
		if authFresh(b, time.Now()) {
			a.mu.Unlock()
			return nil
		}
		if a.active == 0 && !a.refreshing {
			a.refreshing = true
			a.mu.Unlock()
			break
		}
		a.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	defer func() {
		a.mu.Lock()
		a.refreshing = false
		if err != nil {
			a.poisoned = true
		}
		a.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	home := filepath.Join(root, ".auth-maintenance", "home")
	work := filepath.Join(root, ".auth-maintenance", "workspace")
	for _, p := range []string{home, work} {
		if err = os.MkdirAll(p, 0700); err != nil {
			return
		}
	}
	b, err := os.ReadFile(a.Path)
	if err != nil {
		return
	}
	authPath := filepath.Join(home, "auth.json")
	if err = atomicWrite(authPath, b, 0600); err != nil {
		return
	}
	defer os.Remove(authPath)
	args := []string{"-c", `cli_auth_credentials_store="file"`, "-c", `features.shell_tool=false`, "app-server", "--stdio"}
	binary, argv, cleanup, err := l.Build(ctx, home, work, args)
	if err != nil {
		return
	}
	defer cleanup()
	cmd := exec.CommandContext(ctx, binary, argv...)
	cmd.Env = cleanEnv(os.Environ())
	setupProcess(cmd)
	in, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	defer in.Close()
	out, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	cmd.Stderr = &boundedBuffer{limit: 8192}
	if err = cmd.Start(); err != nil {
		return
	}
	encoder := json.NewEncoder(in)
	send := func(v any) error { return encoder.Encode(v) }
	if err = send(map[string]any{"id": 0, "method": "initialize", "params": map[string]any{"clientInfo": map[string]string{"name": "trading_auth_coordinator", "version": "1.0.0"}}}); err != nil {
		cancel()
		_ = cmd.Wait()
		return
	}
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 4096), 128*1024)
	refreshed := false
	for scanner.Scan() {
		var event struct {
			ID     *int            `json:"id"`
			Error  json.RawMessage `json:"error"`
			Result struct {
				Account *struct {
					Type string `json:"type"`
				} `json:"account"`
			} `json:"result"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil || event.ID == nil {
			continue
		}
		if len(event.Error) > 0 && string(event.Error) != "null" {
			err = &Error{"auth_required", "official CLI account refresh rejected"}
			break
		}
		if *event.ID == 0 {
			_ = send(map[string]any{"method": "initialized", "params": map[string]any{}})
			if err = send(map[string]any{"id": 1, "method": "account/read", "params": map[string]bool{"refreshToken": true}}); err != nil {
				break
			}
		}
		if *event.ID == 1 {
			refreshed = event.Result.Account != nil && event.Result.Account.Type == "chatgpt"
			break
		}
	}
	_ = in.Close()
	if !refreshed {
		cancel()
	}
	waitErr := cmd.Wait()
	if err != nil {
		return
	}
	if waitErr != nil || !refreshed {
		return &Error{"auth_required", "managed ChatGPT refresh did not complete"}
	}
	newCache, err := os.ReadFile(authPath)
	if err != nil {
		return
	}
	if !authFresh(newCache, time.Now()) {
		return &Error{"auth_required", "refreshed login does not meet runner freshness requirements"}
	}
	return atomicWrite(a.Path, newCache, 0600)
}
