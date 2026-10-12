package actors

import (
	"encoding/json"
	"fmt"
	"github.com/caglarkc/trading/internal/domain"
	"os"
	"strings"
)

type Personas struct {
	Version       string            `json:"version"`
	CommonVersion string            `json:"common_version"`
	Common        string            `json:"common"`
	Items         map[string]string `json:"personas"`
}

func LoadPersonas(path string) (Personas, error) {
	var p Personas
	b, e := os.ReadFile(path)
	if e != nil {
		return p, e
	}
	if e = json.Unmarshal(b, &p); e != nil {
		return p, e
	}
	if p.Version != "persona-v1" || p.Common == "" || len(p.Items) != 8 {
		return p, fmt.Errorf("persona_config_invalid")
	}
	for _, code := range domain.WalletCodes {
		if p.Items[code] == "" {
			return p, fmt.Errorf("persona_missing")
		}
	}
	for _, value := range append([]string{p.Common}, mapValues(p.Items)...) {
		for _, bad := range []string{"simülasyon", "sanal hesap", "diğer cüzdan", "800.000", "%10", "benchmark"} {
			if strings.Contains(strings.ToLower(value), bad) {
				return p, fmt.Errorf("experiment_metadata_in_persona")
			}
		}
	}
	return p, nil
}
func mapValues(m map[string]string) []string {
	out := []string{}
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
