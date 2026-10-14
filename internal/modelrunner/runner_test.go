package modelrunner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeLauncher struct{ mode string }

func (f fakeLauncher) Build(_ context.Context, _, _ string, _ []string) (string, []string, func(), error) {
	exe, err := os.Executable()
	return exe, []string{"-test.run=TestCLIHelper", "--", f.mode}, func() {}, err
}
func TestCLIHelper(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "--" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	if mode == "hang" {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
	fmt.Println(`{"type":"thread.started","thread_id":"00000000-0000-0000-0000-000000000001"}`)
	if mode == "forbidden" {
		fmt.Println(`{"type":"item.completed","item":{"type":"command_execution"}}`)
		os.Exit(0)
	}
	if mode == "item_error" {
		fmt.Println(`{"type":"item.completed","item":{"type":"error","message":"upstream schema rejected"}}`)
		fmt.Println(`{"type":"turn.failed","error":{"message":"operation cancelled"}}`)
		os.Exit(0)
	}
	if mode == "auth_error" {
		fmt.Println(`{"type":"item.completed","item":{"type":"error","message":"Reconnecting... 2/5 (workspace routing discovery unauthorized (401))"}}`)
		fmt.Println(`{"type":"turn.failed","error":{"message":"operation cancelled"}}`)
		os.Exit(0)
	}
	if mode == "reconnect_ok" {
		fmt.Println(`{"type":"item.completed","item":{"type":"error","message":"Reconnecting... 2/5 (temporary transport failure)"}}`)
	}
	output := `{"ok":true}`
	if mode == "bad" {
		output = `{"extra":true}`
	}
	b, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]any{"type": "agent_message", "text": output}})
	fmt.Println(string(b))
	fmt.Println(`{"type":"turn.completed","usage":{"input_tokens":1}}`)
	os.Exit(0)
}
func TestRunnerPreservesCLIItemError(t *testing.T) {
	r := New(t.TempDir(), &Auth{Path: testAuth(t)}, fakeLauncher{"item_error"})
	got, err := r.Run(context.Background(), Request{Wallet: "A", Prompt: "test"})
	e, ok := err.(*Error)
	if !ok || e.Code != "model_error" || e.Detail != "upstream schema rejected" {
		t.Fatal("original CLI error was masked", err)
	}
	if len(got.Tools) != 0 {
		t.Fatal("CLI error counted as a tool", got.Tools)
	}
}
func TestRunnerClassifiesRoutingAuthFailure(t *testing.T) {
	r := New(t.TempDir(), &Auth{Path: testAuth(t)}, fakeLauncher{"auth_error"})
	_, err := r.Run(context.Background(), Request{Wallet: "A", Prompt: "test"})
	e, ok := err.(*Error)
	if !ok || e.Code != "auth_required" || !strings.Contains(e.Detail, "unauthorized (401)") {
		t.Fatal("authentication failure was masked", err)
	}
}
func TestRunnerAllowsRecoveredReconnect(t *testing.T) {
	r := New(t.TempDir(), &Auth{Path: testAuth(t)}, fakeLauncher{"reconnect_ok"})
	got, err := r.Run(context.Background(), Request{Wallet: "A", Prompt: "test"})
	if err != nil || string(got.Output) != `{"ok":true}` {
		t.Fatal("recovered transport failure rejected", got, err)
	}
}
func testAuth(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "canonical.json")
	jwt := "x." + base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, time.Now().Add(time.Hour).Unix()))) + ".x"
	b, _ := json.Marshal(map[string]any{"tokens": map[string]string{"access_token": jwt}, "last_refresh": time.Now().UTC().Format(time.RFC3339)})
	if err := os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func TestRunnerSchemaAndCleanup(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ok":{"type":"boolean"}},"required":["ok"]}`)
	for _, mode := range []string{"ok", "bad", "forbidden", "hang"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			r := New(root, &Auth{Path: testAuth(t)}, fakeLauncher{mode})
			q := Request{Wallet: "A", Prompt: "test", Schema: schema}
			if mode == "hang" {
				q.Deadline = time.Now().Add(100 * time.Millisecond)
			}
			got, err := r.Run(context.Background(), q)
			if mode == "ok" {
				if err != nil || got.SessionID == "" {
					t.Fatal(got, err)
				}
			} else if err == nil {
				t.Fatal("expected rejection", mode)
			}
			if _, err := os.Stat(filepath.Join(root, "A", "codex-home", "auth.json")); !os.IsNotExist(err) {
				t.Fatal("private auth snapshot retained")
			}
		})
	}
}
func TestWalletTraversalAndExplicitSession(t *testing.T) {
	r := New(t.TempDir(), &Auth{Path: testAuth(t)}, fakeLauncher{"ok"})
	for _, q := range []Request{{Wallet: "../B"}, {Wallet: "A", SessionID: "--last"}, {Wallet: "A", SessionID: "00000000-0000-0000-0000-000000000002", Prompt: "test"}} {
		if _, err := r.Run(context.Background(), q); err == nil {
			t.Fatal("unsafe/mismatched scope accepted", q)
		}
	}
}
func TestAuthRotationConflict(t *testing.T) {
	a := &Auth{Path: testAuth(t)}
	home1, home2 := t.TempDir(), t.TempDir()
	l1, err := a.acquire(context.Background(), home1)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := a.acquire(context.Background(), home2)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(l1.path, []byte(`{"new":"one"}`), 0600)
	_ = os.WriteFile(l2.path, []byte(`{"new":"two"}`), 0600)
	if err = l1.release(); err != nil {
		t.Fatal(err)
	}
	if err = l2.release(); err == nil {
		t.Fatal("divergent refresh accepted")
	}
	if _, err = a.acquire(context.Background(), t.TempDir()); err == nil {
		t.Fatal("poisoned auth continued")
	}
}
func TestDockerIsolation(t *testing.T) {
	d := Docker{Image: "image@sha256:fixed", Packages: "/cli", UID: 1001, GID: 1001}
	_, args, _, err := d.Build(context.Background(), "/private/A/home", "/private/A/work", []string{"exec"})
	if err != nil {
		t.Fatal(err)
	}
	command := strings.Join(args, " ")
	for _, want := range []string{"--read-only", "--cap-drop=ALL", "no-new-privileges", "src=/private/A/home", "src=/private/A/work", "src=/cli"} {
		if !strings.Contains(command, want) {
			t.Fatal(want)
		}
	}
	for _, bad := range []string{"docker.sock", "/private/B", "src=/,", "--privileged", "API_KEY"} {
		if strings.Contains(command, bad) {
			t.Fatal("unsafe mount/environment", bad)
		}
	}
}
