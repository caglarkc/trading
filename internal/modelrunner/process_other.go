//go:build !unix

package modelrunner

import "os/exec"

func setupProcess(cmd *exec.Cmd) {}
