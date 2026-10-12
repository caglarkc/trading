package modelrunner

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Each model process sees only its own home/workspace and the public CLI
// distribution. It receives no host root, repository, Docker socket or SSH key.
type Docker struct {
	Binary, Image, Packages string
	UID, GID                int
	Network                 string
}

func (d Docker) Build(ctx context.Context, home, work string, args []string) (string, []string, func(), error) {
	if d.Binary == "" {
		d.Binary = "docker"
	}
	if d.Image == "" || d.Packages == "" {
		return "", nil, nil, fmt.Errorf("Docker CLI image/packages required")
	}
	name := fmt.Sprintf("trading-gpt-%d", time.Now().UnixNano())
	argv := []string{"run", "--rm", "-i", "--name", name, "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--pids-limit=128", "--cpus=0.75", "--memory=768m", "--user", fmt.Sprintf("%d:%d", d.UID, d.GID), "--tmpfs", "/tmp:rw,nosuid,nodev,size=128m,mode=1777", "--mount", "type=bind,src=" + home + ",dst=/state", "--mount", "type=bind,src=" + work + ",dst=/workspace", "--mount", "type=bind,src=" + d.Packages + ",dst=/opt/codex,readonly", "--env", "CODEX_HOME=/state", "--env", "HOME=/state", "--workdir", "/workspace"}
	if d.Network != "" {
		argv = append(argv, "--network", d.Network)
	}
	argv = append(argv, d.Image, "/opt/codex/bin/codex")
	argv = append(argv, args...)
	cleanup := func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = exec.CommandContext(c, d.Binary, "rm", "-f", name).Run()
	}
	return d.Binary, argv, cleanup, nil
}
