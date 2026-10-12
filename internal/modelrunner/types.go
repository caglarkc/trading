package modelrunner

import (
	"context"
	"encoding/json"
	"time"
)

const Model = "gpt-6.1-sol"
const Reasoning = "low"

type Request struct {
	Wallet    string          `json:"wallet"`
	SessionID string          `json:"session_id,omitempty"`
	Prompt    string          `json:"prompt"`
	Schema    json.RawMessage `json:"schema"`
	Web       bool            `json:"web"`
	Deadline  time.Time       `json:"deadline"`
}

type Result struct {
	SessionID  string          `json:"session_id"`
	Output     json.RawMessage `json:"output"`
	DurationMS int64           `json:"duration_ms"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt time.Time       `json:"finished_at"`
	Tools      []string        `json:"tools"`
	Usage      json.RawMessage `json:"usage,omitempty"`
	OutputHash string          `json:"output_hash"`
	Model      string          `json:"model"`
	Reasoning  string          `json:"reasoning"`
}

type Launcher interface {
	Build(context.Context, string, string, []string) (string, []string, func(), error)
}

type Error struct {
	Code   string
	Detail string
}

func (e *Error) Error() string { return e.Code + ": " + e.Detail }
