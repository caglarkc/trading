package modelrunner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCoordinatorAndOrphanAuth(t *testing.T) {
	root := t.TempDir()
	unlock, e := CoordinatorLock(root)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock()
	if release, e := CoordinatorLock(root); e == nil {
		release()
		t.Fatal("second coordinator acquired lock")
	}
	auth := filepath.Join(root, "canonical")
	os.WriteFile(auth, []byte(`{"tokens":"canonical"}`), 0600)
	p := filepath.Join(root, "wallet", "codex-home", "auth.json")
	os.MkdirAll(filepath.Dir(p), 0700)
	os.WriteFile(p, []byte(`{"tokens":"changed"}`), 0600)
	r := New(root, &Auth{Path: auth}, nil)
	if r.Recover(context.Background()) == nil {
		t.Fatal("orphan rotation silently discarded")
	}
	os.WriteFile(p, []byte(`{"tokens":"canonical"}`), 0600)
	if e = r.Recover(context.Background()); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(p); !os.IsNotExist(e) {
		t.Fatal("auth lease not cleared")
	}
}
func TestSanitizedEventKeepsValidJSON(t *testing.T) {
	b := safeEvent([]byte(`{"type":"item.completed","item":{"text":"` + string(make([]byte, 0)) + `hello","access_token":"secret"}}`))
	if !json.Valid(b) {
		t.Fatal(string(b))
	}
	var m map[string]any
	json.Unmarshal(b, &m)
	if m["item"].(map[string]any)["access_token"] == "secret" {
		t.Fatal("token exposed")
	}
}
