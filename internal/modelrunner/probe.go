package modelrunner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type ProbeReport struct {
	Calls               []Result `json:"calls"`
	PrivateRecall       bool     `json:"private_recall"`
	CrossResumeRejected bool     `json:"cross_resume_rejected"`
	MaxConcurrent       int      `json:"max_concurrent"`
}

func Probe(ctx context.Context, r *Runner) (report ProbeReport, err error) {
	schema := json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"account":{"type":"string"},"marker":{"type":"string"},"other_marker":{"type":"null"}},"required":["account","marker","other_marker"]}`)
	names := []string{"probe-A", "probe-B"}
	markers := make([]string, 2)
	sessions := make([]string, 2)
	for i := range 2 {
		b := make([]byte, 16)
		if _, err = rand.Read(b); err != nil {
			return
		}
		markers[i] = hex.EncodeToString(b)
	}
	var mu sync.Mutex
	for round := range 2 {
		var wg sync.WaitGroup
		for i := range 2 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				prompt := fmt.Sprintf("Read-only private account %s probe. No tools/files/search. Recall this conversation's private marker. Return only schema JSON, account=%s, marker=your private marker, other_marker=null.", names[i], names[i])
				if round == 0 {
					prompt = fmt.Sprintf("Read-only private account %s probe. No tools/files/search. Remember your private marker %s in this conversation only. Return only schema JSON, account=%s, marker=that marker, other_marker=null.", names[i], markers[i], names[i])
				}
				value, e := r.Run(ctx, Request{Wallet: names[i], SessionID: sessions[i], Prompt: prompt, Schema: schema})
				mu.Lock()
				defer mu.Unlock()
				if e != nil {
					if err == nil {
						err = e
					}
					return
				}
				var answer struct {
					Account, Marker string
					Other           *string `json:"other_marker"`
				}
				e = json.Unmarshal(value.Output, &answer)
				if e != nil || answer.Account != names[i] || answer.Marker != markers[i] || answer.Other != nil {
					err = fmt.Errorf("private recall failed for %s", names[i])
					return
				}
				sessions[i] = value.SessionID
				// Retain measured metadata, never the private marker in the public report.
				value.Output = json.RawMessage(`{"private_recall":"passed"}`)
				report.Calls = append(report.Calls, value)
			}(i)
		}
		wg.Wait()
		if err != nil {
			return
		}
	}
	report.PrivateRecall = true
	type point struct {
		at    time.Time
		delta int
	}
	points := []point{}
	for _, call := range report.Calls {
		points = append(points, point{call.StartedAt, 1}, point{call.FinishedAt, -1})
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].at.Equal(points[j].at) {
			return points[i].delta < points[j].delta
		}
		return points[i].at.Before(points[j].at)
	})
	active := 0
	for _, p := range points {
		active += p.delta
		report.MaxConcurrent = max(report.MaxConcurrent, active)
	}
	if report.MaxConcurrent != 2 {
		return report, fmt.Errorf("two concurrent calls were not demonstrated")
	}
	_, crossErr := r.Run(ctx, Request{Wallet: names[1], SessionID: sessions[0], Prompt: "No tools. Return JSON.", Schema: schema})
	report.CrossResumeRejected = crossErr != nil && strings.Contains(crossErr.Error(), "no rollout found")
	if !report.CrossResumeRejected {
		err = fmt.Errorf("cross-home resume was not rejected as expected")
	}
	return
}
