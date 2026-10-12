package modelrunner

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var walletPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,63}$`)
var sessionPattern = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9-]{27}$`)

type Runner struct {
	Root     string
	Auth     *Auth
	Launcher Launcher
	slots    chan struct{}
	mu       sync.Mutex
	busy     map[string]bool
}

func New(root string, auth *Auth, l Launcher) *Runner {
	return &Runner{Root: root, Auth: auth, Launcher: l, slots: make(chan struct{}, 2), busy: map[string]bool{}}
}

func (r *Runner) Run(ctx context.Context, q Request) (result Result, err error) {
	if !walletPattern.MatchString(q.Wallet) {
		return result, &Error{"invalid_wallet", "invalid wallet scope"}
	}
	if q.SessionID != "" && !sessionPattern.MatchString(q.SessionID) {
		return result, &Error{"invalid_session", "invalid explicit session ID"}
	}
	if len(q.Prompt) > 1024*1024 || len(q.Schema) > 256*1024 {
		return result, &Error{"payload_too_large", "runner input exceeds bound"}
	}
	if q.Deadline.IsZero() {
		q.Deadline = time.Now().Add(6 * time.Minute)
	}
	ctx, cancel := context.WithDeadline(ctx, q.Deadline)
	defer cancel()
	select {
	case r.slots <- struct{}{}:
	case <-ctx.Done():
		return result, ctx.Err()
	}
	defer func() { <-r.slots }()
	r.mu.Lock()
	if r.busy[q.Wallet] {
		r.mu.Unlock()
		return result, &Error{"wallet_busy", "one writer per wallet"}
	}
	r.busy[q.Wallet] = true
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.busy, q.Wallet); r.mu.Unlock() }()
	if err = r.Auth.EnsureFresh(ctx, r.Root, r.Launcher); err != nil {
		return result, err
	}
	root, err := filepath.Abs(filepath.Join(r.Root, q.Wallet))
	if err != nil {
		return result, err
	}
	home, work := filepath.Join(root, "codex-home"), filepath.Join(root, "workspace")
	for _, p := range []string{home, work} {
		if err = os.MkdirAll(p, 0700); err != nil {
			return result, err
		}
	}
	var schema *jsonschema.Schema
	if len(q.Schema) > 0 {
		var doc any
		if err = json.Unmarshal(q.Schema, &doc); err != nil {
			return result, err
		}
		compiler := jsonschema.NewCompiler()
		if err = compiler.AddResource("decision.json", doc); err != nil {
			return result, err
		}
		schema, err = compiler.Compile("decision.json")
		if err != nil {
			return result, err
		}
		if err = atomicWrite(filepath.Join(work, "output.schema.json"), q.Schema, 0600); err != nil {
			return result, err
		}
	}
	lease, err := r.Auth.acquire(ctx, home)
	if err != nil {
		return result, err
	}
	defer func() { err = joinError(err, lease.release()) }()
	args := []string{"exec", "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check", "--json", "--sandbox", "read-only", "-m", Model, "-c", `model_reasoning_effort="` + Reasoning + `"`, "-c", `features.shell_tool=false`, "-c", `features.unified_exec=false`, "-c", `forced_login_method="chatgpt"`, "-c", `cli_auth_credentials_store="file"`}
	web := "disabled"
	if q.Web {
		web = "live"
	}
	args = append(args, "-c", `web_search="`+web+`"`)
	if q.SessionID != "" {
		args = append(args, "resume", q.SessionID)
	}
	if schema != nil {
		args = append(args, "--output-schema", "/workspace/output.schema.json")
	}
	args = append(args, "-")
	binary, argv, cleanup, err := r.Launcher.Build(ctx, home, work, args)
	if err != nil {
		return result, err
	}
	defer cleanup()
	cmd := exec.CommandContext(ctx, binary, argv...)
	cmd.Stdin = strings.NewReader(q.Prompt)
	cmd.Env = cleanEnv(os.Environ())
	setupProcess(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}
	stderr := &boundedBuffer{limit: 64 * 1024}
	cmd.Stderr = stderr
	start := time.Now()
	if err = cmd.Start(); err != nil {
		return result, err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	var output string
	var protocolError error
	total := 0
	result = Result{Model: Model, Reasoning: Reasoning, Tools: []string{}}
	result.StartedAt = start.UTC()
	for scanner.Scan() {
		total += len(scanner.Bytes())
		if total > 8*1024*1024 {
			cancel()
			protocolError = &Error{"output_too_large", "JSONL exceeds 8 MiB"}
			break
		}
		var e struct {
			Type     string                      `json:"type"`
			ThreadID string                      `json:"thread_id"`
			Usage    json.RawMessage             `json:"usage"`
			Item     struct{ Type, Text string } `json:"item"`
			Message  string                      `json:"message"`
		}
		if json.Unmarshal(scanner.Bytes(), &e) != nil {
			protocolError = &Error{"invalid_jsonl", "malformed CLI event"}
			continue
		}
		switch e.Type {
		case "thread.started":
			result.SessionID = e.ThreadID
		case "turn.completed":
			result.Usage = e.Usage
		case "error", "turn.failed":
			protocolError = &Error{"model_error", redact(e.Message)}
		}
		if e.Type == "item.completed" {
			if e.Item.Type == "agent_message" {
				output = e.Item.Text
			} else if e.Item.Type != "" {
				result.Tools = append(result.Tools, e.Item.Type)
				if e.Item.Type != "web_search" {
					cancel()
					protocolError = &Error{"forbidden_tool", "unexpected tool type " + e.Item.Type}
				}
			}
		}
	}
	readErr := scanner.Err()
	waitErr := cmd.Wait()
	result.DurationMS = time.Since(start).Milliseconds()
	result.FinishedAt = time.Now().UTC()
	if ctx.Err() != nil {
		return result, &Error{"session_deadline", "CLI deadline or cancellation"}
	}
	if protocolError != nil {
		return result, protocolError
	}
	if readErr != nil {
		return result, readErr
	}
	if waitErr != nil {
		return result, &Error{"model_process_failed", redact(stderr.String())}
	}
	if !sessionPattern.MatchString(result.SessionID) {
		return result, &Error{"invalid_session", "CLI did not return session ID"}
	}
	if q.SessionID != "" && result.SessionID != q.SessionID {
		return result, &Error{"session_mismatch", "explicit resume changed session"}
	}
	if len(output) > 256*1024 || !json.Valid([]byte(output)) {
		return result, &Error{"invalid_output", "final response is not bounded JSON"}
	}
	var value any
	_ = json.Unmarshal([]byte(output), &value)
	if schema != nil {
		if e := schema.Validate(value); e != nil {
			return result, &Error{"invalid_output", e.Error()}
		}
	}
	result.Output = json.RawMessage(output)
	hash := sha256.Sum256([]byte(output))
	result.OutputHash = hex.EncodeToString(hash[:])
	return result, nil
}

func cleanEnv(env []string) []string {
	out := []string{}
	for _, v := range env {
		key, _, _ := strings.Cut(v, "=")
		if strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "CODEX_") {
			continue
		}
		out = append(out, v)
	}
	return out
}
func joinError(a, b error) error {
	if a != nil {
		return a
	}
	return b
}
func redact(s string) string {
	s = regexp.MustCompile(`[A-Za-z0-9_-]{15,}\.[A-Za-z0-9_-]{15,}\.[A-Za-z0-9_-]{10,}`).ReplaceAllString(s, "[redacted-token]")
	if strings.Contains(s, "access_token") || strings.Contains(s, "refresh_token") || strings.Contains(s, "sk-") {
		return "credential-bearing diagnostic suppressed"
	}
	if len(s) > 1200 {
		s = s[:1200]
	}
	return s
}

type boundedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.Len()
	if remaining > 0 {
		_, _ = b.Buffer.Write(p[:min(remaining, n)])
	}
	return n, nil
}

var _ io.Writer = (*boundedBuffer)(nil)
