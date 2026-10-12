package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caglarkc/trading/internal/marketdata"
	"os"
	"time"
)

func main() {
	probe := flag.Bool("probe", false, "collect public source acceptance evidence")
	assets := flag.String("assets", "config/assets.json", "canonical asset registry")
	chrome := flag.String("chrome", "/usr/bin/chromium-browser", "headless Chromium path")
	symbol := flag.String("symbol", "", "one symbol or default seven-symbol probe")
	flag.Parse()
	if !*probe {
		fmt.Fprintln(os.Stderr, "worker service will be configured after database/jobs phase")
		os.Exit(1)
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
		snapshot := reader.Collect(ctx, a, false)
		cancel()
		if e = encoder.Encode(snapshot); e != nil {
			fail(e)
		}
	}
}
func fail(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
