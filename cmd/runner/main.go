package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caglarkc/trading/internal/actors"
	"github.com/caglarkc/trading/internal/bridge"
	"github.com/caglarkc/trading/internal/modelrunner"
	"github.com/caglarkc/trading/internal/research"
	"github.com/caglarkc/trading/internal/store"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mcp := flag.Bool("mcp", false, "scoped stdio MCP bridge")
	socket := flag.String("socket", "/workspace/bridge.sock", "private bridge socket")
	toolsBinary := flag.String("tools-binary", "", "read-only scoped tool binary")
	environment := flag.String("environment", "staging", "owned container environment label")
	personasPath := flag.String("personas", "/opt/trading/staging-src/config/personas.json", "trusted persona config")
	walletProbe := flag.Int("wallet-probe", 0, "real private wallet/receipt acceptance: 2 or 8")
	service := flag.Bool("service", false, "persistent model job coordinator")
	researchProbe := flag.Bool("research-probe", false, "real shared research pipeline acceptance")
	root := flag.String("root", "/var/lib/trading/staging/wallets", "private session root")
	auth := flag.String("auth", "/var/lib/trading/.codex/auth.json", "trusted canonical login cache")
	packages := flag.String("packages", "/var/lib/trading/.codex/packages/standalone/releases/0.162.1-x86_64-unknown-linux-musl", "official CLI package")
	image := flag.String("image", "trading-codex:0.162.1", "isolated CLI runtime image")
	uid := flag.Int("uid", os.Getuid(), "CLI container UID")
	gid := flag.Int("gid", os.Getgid(), "CLI container GID")
	network := flag.String("network", "trading-staging-runner", "CLI network")
	probe := flag.Bool("probe", false, "real isolated two-wallet connection acceptance")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *mcp {
		if e := bridge.MCP(ctx, *socket, os.Stdin, os.Stdout); e != nil {
			fail(e)
		}
		return
	}
	runner := modelrunner.New(*root, &modelrunner.Auth{Path: *auth}, modelrunner.Docker{Image: *image, Packages: *packages, UID: *uid, GID: *gid, Network: *network, ToolsBinary: *toolsBinary, Environment: *environment})
	if *service || *researchProbe || *walletProbe > 0 {
		db, e := store.Open(ctx, os.Getenv("DATABASE_WORKER_URL"))
		if e != nil {
			fail(e)
		}
		defer db.Close()
		svc := &research.Service{Pool: db.Pool, Runner: runner, Collector: os.Getenv("COLLECTOR_URL")}
		p, e := actors.LoadPersonas(*personasPath)
		if e != nil {
			fail(e)
		}
		actor := &actors.Service{Store: db, Runner: runner, Personas: p, Root: *root, API: os.Getenv("API_URL"), Collector: os.Getenv("COLLECTOR_URL"), Secret: []byte(os.Getenv("AGENT_SIGNING_SECRET"))}
		svc.WalletHandler = actor.Run
		if *walletProbe > 0 {
			if *toolsBinary == "" {
				fail(fmt.Errorf("tools-binary required for wallet probe"))
			}
			if e = actor.Probe(ctx, *walletProbe); e != nil {
				fail(e)
			}
			return
		}
		if e = svc.Loop(ctx, *root, *researchProbe); e != nil {
			fail(e)
		}
		return
	}
	if *probe {
		report, err := modelrunner.Probe(ctx, runner)
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
	result, err := runner.Run(ctx, req)
	if err != nil {
		fail(err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
