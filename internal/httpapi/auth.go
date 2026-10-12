package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Claims struct {
	Wallet   string `json:"wallet"`
	Decision string `json:"decision"`
	Job      string `json:"job"`
	Expires  int64  `json:"exp"`
}

func Sign(secret []byte, c Claims) string {
	b, _ := json.Marshal(c)
	part := base64.RawURLEncoding.EncodeToString(b)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(part))
	return part + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func Verify(secret []byte, token string, now time.Time) (Claims, error) {
	var c Claims
	p, s, ok := strings.Cut(token, ".")
	if !ok || len(token) > 2048 {
		return c, fmt.Errorf("invalid_token")
	}
	b, e := base64.RawURLEncoding.DecodeString(s)
	if e != nil {
		return c, fmt.Errorf("invalid_token")
	}
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(p))
	if !hmac.Equal(b, m.Sum(nil)) {
		return c, fmt.Errorf("invalid_token")
	}
	body, e := base64.RawURLEncoding.DecodeString(p)
	if e != nil {
		return c, fmt.Errorf("invalid_token")
	}
	if e = json.Unmarshal(body, &c); e != nil {
		return c, fmt.Errorf("invalid_token")
	}
	if c.Expires <= now.Unix() || c.Expires > now.Add(15*time.Minute).Unix() {
		return c, fmt.Errorf("expired_or_excessive_token")
	}
	for _, id := range []string{c.Wallet, c.Decision, c.Job} {
		if _, e := uuid.Parse(id); e != nil {
			return c, fmt.Errorf("invalid_scope")
		}
	}
	return c, nil
}
func EqualSecret(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
