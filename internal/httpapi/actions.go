package httpapi

import (
	"encoding/json"
	"github.com/caglarkc/trading/internal/domain"
	"io"
	"net/http"
	"time"
)

func (s *Server) orderAction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	var x domain.Action
	if e := d.Decode(&x); e != nil {
		fail(w, 400, "invalid_action_json")
		return
	}
	if e := d.Decode(new(any)); e != io.EOF {
		fail(w, 400, "trailing_json")
		return
	}
	c := ClaimsFrom(r.Context())
	v, e := s.Agent.ApplyAction(r.Context(), c.Wallet, c.Decision, x, time.Now().UTC(), s.Calendar, s.Fees)
	if e != nil {
		code := e.Error()
		status := 409
		if code == "decision_not_found" || code == "decision_not_actionable" {
			status = 403
		}
		fail(w, status, code)
		return
	}
	respond(w, 200, v)
}
