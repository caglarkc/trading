package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPFixedToolsAndPrivateSocket(t *testing.T) {
	dir, e := os.MkdirTemp("/tmp", "trading-mcp-")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	socket := filepath.Join(dir, "bridge.sock")
	l, e := net.Listen("unix", socket)
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/tool" {
			t.Error("unexpected route")
		}
		w.Write([]byte(`{"available_cash_usd":"100000"}`))
	})}
	go server.Serve(l)
	defer server.Close()
	input := strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2025-06-18\"}}\n{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/list\"}\n{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"account\",\"arguments\":{}}}\n")
	var out bytes.Buffer
	if e = MCP(context.Background(), socket, input, &out); e != nil {
		t.Fatal(e)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatal(out.String())
	}
	var response map[string]any
	if e = json.Unmarshal([]byte(lines[2]), &response); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(lines[2], "100000") {
		t.Fatal(response)
	}
	for _, tool := range Tools() {
		if strings.Contains(tool.Name, "submit") || strings.Contains(tool.Name, "operator") || strings.Contains(tool.Name, "shell") {
			t.Fatal("unsafe tool", tool.Name)
		}
	}
}
func TestPrivateToolRejectsWalletOverride(t *testing.T) {
	s := Scope{}
	for _, body := range []string{`{"Name":"operator","Arguments":{}}`, `{"Name":"account","Arguments":{"wallet_id":"other"}}`, `{"Name":"account","Arguments":{"path":"/etc/trading"}}`} {
		req, _ := http.NewRequest("POST", "http://bridge/tool", strings.NewReader(body))
		w := httptest.NewRecorder()
		s.tool(w, req)
		if w.Code == 200 {
			t.Fatal("unsafe request accepted", body)
		}
	}
}
