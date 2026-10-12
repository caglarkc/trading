package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caglarkc/trading/internal/modelrunner"
	"os"
)

func main() {
	root := flag.String("root", "/var/lib/trading/staging/wallets", "private session root")
	auth := flag.String("auth", "/var/lib/trading/.codex/auth.json", "trusted canonical login cache")
	packages := flag.String("packages", "/var/lib/trading/.codex/packages/standalone/releases/0.162.1-x86_64-unknown-linux-musl", "official CLI package")
	image := flag.String("image", "trading-codex:0.162.1", "isolated CLI runtime image")
	uid := flag.Int("uid", os.Getuid(), "CLI container UID")
	gid := flag.Int("gid", os.Getgid(), "CLI container GID")
	network := flag.String("network", "trading-staging-runner", "CLI network")
	probe := flag.Bool("probe", false, "real isolated two-wallet connection acceptance")
	flag.Parse()
	runner := modelrunner.New(*root, &modelrunner.Auth{Path: *auth}, modelrunner.Docker{Image: *image, Packages: *packages, UID: *uid, GID: *gid, Network: *network})
	if *probe {
		report, err := modelrunner.Probe(context.Background(), runner)
		if err != nil {
			fail(err)
		}
		if err = json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fail(err)
		}
		return
	}
	var req modelrunner.Request
	dec := json.NewDecoder(os.Stdin)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fail(err)
	}
	result, err := runner.Run(context.Background(), req)
	if err != nil {
		fail(err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
