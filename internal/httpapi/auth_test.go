package httpapi

import (
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestScope(t *testing.T) {
	now := time.Now()
	c := Claims{uuid.NewString(), uuid.NewString(), uuid.NewString(), now.Add(time.Minute).Unix()}
	token := Sign([]byte("fixture-secret"), c)
	got, e := Verify([]byte("fixture-secret"), token, now)
	if e != nil || got != c {
		t.Fatal(got, e)
	}
	for _, tc := range []struct {
		key, token string
		at         time.Time
	}{{"wrong", token, now}, {"fixture-secret", token + "a", now}, {"fixture-secret", token, now.Add(time.Minute)}} {
		if _, e := Verify([]byte(tc.key), tc.token, tc.at); e == nil {
			t.Fatal("bad token accepted")
		}
	}
}
