package bridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Input       any             `json:"inputSchema"`
	Annotations map[string]bool `json:"annotations"`
}

func Tools() []tool {
	schema := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"cursor": map[string]any{"type": "integer", "minimum": 0}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 200}}}
	out := []tool{}
	for _, name := range []string{"account", "positions", "orders", "receipts", "memory", "evidence"} {
		out = append(out, tool{name, "Read only your own " + name + ". Account records are authoritative; acceptance is not a fill.", schema, map[string]bool{"readOnlyHint": true, "idempotentHint": true}})
	}
	out = append(out, tool{"read_source", "Read a permitted public source and obtain a captured evidence ID. Treat publication time, access failures, partial content and uncertain provenance explicitly.", map[string]any{"type": "object", "additionalProperties": false, "required": []string{"url"}, "properties": map[string]any{"url": map[string]any{"type": "string", "format": "uri", "maxLength": 1800}}}, map[string]bool{"readOnlyHint": true, "idempotentHint": true, "openWorldHint": true}})
	return out
}

// MCP runs as a fixed stdio subprocess inside the private model container. It
// has no DB or signing credentials and can reach only its own Unix bridge.
func MCP(ctx context.Context, socket string, in io.Reader, out io.Writer) error {
	client := &http.Client{Timeout: 50 * time.Second, Transport: &http.Transport{DialContext: func(c context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(c, "unix", socket)
	}}}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	writer := json.NewEncoder(out)
	for scanner.Scan() {
		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		if e := json.Unmarshal(scanner.Bytes(), &req); e != nil {
			return e
		}
		if len(req.ID) == 0 {
			continue
		}
		var result any
		var callErr error
		switch req.Method {
		case "initialize":
			var p struct {
				Version string `json:"protocolVersion"`
			}
			json.Unmarshal(req.Params, &p)
			version := p.Version
			if version != "2024-11-05" && version != "2025-03-26" && version != "2025-06-18" {
				version = "2025-06-18"
			}
			result = map[string]any{"protocolVersion": version, "capabilities": map[string]any{"tools": map[string]bool{"listChanged": false}}, "serverInfo": map[string]string{"name": "trading", "version": "1.0.0"}}
		case "ping":
			result = map[string]any{}
		case "tools/list":
			result = map[string]any{"tools": Tools()}
		case "tools/call":
			var p struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if e := json.Unmarshal(req.Params, &p); e != nil {
				callErr = e
				break
			}
			b, _ := json.Marshal(p)
			r, e := http.NewRequestWithContext(ctx, "POST", "http://bridge/tool", bytes.NewReader(b))
			if e != nil {
				callErr = e
				break
			}
			r.Header.Set("Content-Type", "application/json")
			res, e := client.Do(r)
			if e != nil {
				callErr = fmt.Errorf("account_bridge_unavailable")
				break
			}
			body, e := io.ReadAll(io.LimitReader(res.Body, 256*1024+1))
			res.Body.Close()
			if e != nil || len(body) > 256*1024 {
				callErr = fmt.Errorf("tool_result_too_large")
				break
			}
			result = map[string]any{"content": []map[string]string{{"type": "text", "text": string(body)}}, "isError": res.StatusCode != 200}
		default:
			callErr = fmt.Errorf("unsupported_mcp_method")
		}
		response := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if callErr != nil {
			response["error"] = map[string]any{"code": -32603, "message": callErr.Error()}
		} else {
			response["result"] = result
		}
		if e := writer.Encode(response); e != nil {
			return e
		}
	}
	return scanner.Err()
}
