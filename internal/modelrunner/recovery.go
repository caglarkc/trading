package modelrunner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// CoordinatorLock is process-wide; native calls and probes share it.
func CoordinatorLock(root string) (func(), error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(root, "runner.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("runner_already_active")
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}
func rootLabel(root string) string {
	p, _ := filepath.Abs(root)
	h := sha256.Sum256([]byte(p))
	return hex.EncodeToString(h[:])
}

// Recover is called only under the coordinator lock. Unknown rotated auth is
// preserved for trusted reconciliation; a stale refresh token is never restored.
func (r *Runner) Recover(ctx context.Context) error {
	if d, ok := r.Launcher.(Docker); ok {
		binary := d.Binary
		if binary == "" {
			binary = "docker"
		}
		if d.Environment == "" {
			return fmt.Errorf("runner_environment_required")
		}
		b, err := exec.CommandContext(ctx, binary, "ps", "-aq", "--filter", "label=trading.runner="+d.Environment, "--filter", "label=trading.root="+rootLabel(r.Root)).Output()
		if err != nil {
			return err
		}
		for _, id := range strings.Fields(string(b)) {
			if err = exec.CommandContext(ctx, binary, "rm", "-f", id).Run(); err != nil {
				return err
			}
		}
	}
	canonical, err := os.ReadFile(r.Auth.Path)
	if err != nil {
		return &Error{"auth_required", "canonical login unavailable"}
	}
	paths, err := filepath.Glob(filepath.Join(r.Root, "*", "codex-home", "auth.json"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if sha256.Sum256(b) != sha256.Sum256(canonical) {
			return &Error{"auth_required", "orphan credential rotation requires trusted reconciliation"}
		}
		if err = os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}

// CheckpointLease prevents DB/session checkpoints from racing private writers.
func CheckpointLease(ctx context.Context, root string, exclusive bool) (func(), error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(root, "checkpoint.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	for {
		if err = syscall.Flock(int(f.Fd()), how|syscall.LOCK_NB); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			f.Close()
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}
