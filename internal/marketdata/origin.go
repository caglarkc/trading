package marketdata

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

// Explicit agency attribution groups syndicated copies under one origin.
func Origin(target, text string) string {
	if len(text) > 3000 {
		text = text[:3000]
	}
	t := strings.ToLower(text)
	if strings.Contains(t, "reuters") {
		return "wire:reuters"
	}
	if strings.Contains(t, "associated press") {
		return "wire:ap"
	}
	if strings.Contains(t, "pr newswire") || strings.Contains(t, "prnewswire") {
		return "wire:prnewswire"
	}
	if strings.Contains(t, "business wire") {
		return "wire:businesswire"
	}
	u, e := url.Parse(target)
	if e != nil {
		return "unknown"
	}
	host := strings.TrimPrefix(u.Hostname(), "www.")
	return "publisher:" + host
}
func TextHash(text string) string {
	h := sha256.Sum256([]byte(strings.Join(strings.Fields(text), " ")))
	return hex.EncodeToString(h[:])
}
