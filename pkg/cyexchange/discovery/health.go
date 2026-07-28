// Status: PROPOSED
// authority_effect: NONE
package discovery

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// Health represents the current stability of the cyxd daemon.
type Health struct {
	IsReady bool `json:"is_ready"`
	IsLive  bool `json:"is_live"`
}

// Prober manages the `/healthz` liveness and readiness states.
type Prober struct {
	isReady atomic.Bool
	isLive  atomic.Bool
}

func NewProber() *Prober {
	p := &Prober{}
	p.isLive.Store(true) // Live as soon as the process starts
	return p
}

// MarkReady indicates the daemon has finished recovering its SQLite log.
func (p *Prober) MarkReady() {
	p.isReady.Store(true)
}

// Handler returns an HTTP handler for health probing.
// Strictly 010: this is structural metadata, not a "production-ready" marketing claim.
func (p *Prober) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := Health{
			IsReady: p.isReady.Load(),
			IsLive:  p.isLive.Load(),
		}

		if !h.IsReady {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(h)
	}
}
