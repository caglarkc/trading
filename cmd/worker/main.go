package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caglarkc/trading/internal/calendar"
	"github.com/caglarkc/trading/internal/domain"
	"github.com/caglarkc/trading/internal/marketdata"
	"github.com/caglarkc/trading/internal/store"
	workservice "github.com/caglarkc/trading/internal/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	renderServer := flag.Bool("render-server", false, "isolated Chromium reader service")
	listen := flag.String("listen", "127.0.0.1:18881", "render/source service listen")
	hosts := flag.String("hosts", "config/source-hosts.json", "permitted public hosts")
	probe := flag.Bool("probe", false, "collect public source acceptance evidence")
	assets := flag.String("assets", "config/assets.json", "canonical asset registry")
	chrome := flag.String("chrome", "/usr/bin/chromium-browser", "headless Chromium path")
	symbol := flag.String("symbol", "", "one symbol or default seven-symbol probe")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *renderServer {
		if e := marketdata.ServeRender(ctx, *listen, *chrome, *hosts); e != nil {
			fail(e)
		}
		return
	}
	if !*probe {
		if os.Getenv("DATABASE_WORKER_URL") == "" {
			fail(fmt.Errorf("DATABASE_WORKER_URL required"))
		}
		db, e := store.Open(ctx, os.Getenv("DATABASE_WORKER_URL"))
		if e != nil {
			fail(e)
		}
		defer db.Close()
		cal, e := calendar.Load(os.Getenv("CALENDAR_FILE"))
		if e != nil {
			fail(e)
		}
		fee, e := domain.LoadFees(os.Getenv("FEES_FILE"))
		if e != nil {
			fail(e)
		}
		reader := marketdata.NewReader("")
		if e = reader.LoadHosts(*hosts); e != nil {
			fail(e)
		}
		reader.RenderURL = os.Getenv("RENDER_URL")
		w := &workservice.Worker{Store: db, Reader: reader, Calendar: cal, Fees: fee, Execution: os.Getenv("EXECUTION_ENABLED") == "true"}
		go func() {
			if e := w.Serve(ctx, *listen); e != nil {
				log.Print("source service stopped")
				stop()
			}
		}()
		if e = w.Loop(ctx); e != nil {
			fail(e)
		}
		return
	}
	b, e := os.ReadFile(*assets)
	if e != nil {
		fail(e)
	}
	list, e := marketdata.DecodeAssets(b)
	if e != nil {
		fail(e)
	}
	reader := marketdata.NewReader(*chrome)
	reader.RenderURL = os.Getenv("RENDER_URL")
	if reader.RenderURL != "" {
		reader.ChromePath = ""
	}
	calFile := os.Getenv("CALENDAR_FILE")
	if calFile == "" {
		calFile = "config/calendar.json"
	}
	cal, e := calendar.Load(calFile)
	if e != nil {
		fail(e)
	}
	encoder := json.NewEncoder(os.Stdout)
	wanted := map[string]bool{"QQQ": true, "TQQQ": true, "EWC": true, "AAPL": true, "JPM": true, "XOM": true, "BRK.B": true}
	for _, a := range list {
		if *symbol != "" && a.Symbol != *symbol {
			continue
		}
		if *symbol == "" && !wanted[a.Symbol] {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		snapshot := reader.Collect(ctx, a, cal.Open(time.Now()))
		cancel()
		if e = encoder.Encode(snapshot); e != nil {
			fail(e)
		}
	}
}
func fail(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
