// Package server is Mímir's live HTTP interface: it lets a browser (or any client)
// talk to the brain. GET / serves the polished book-spine UI (E70 / F53); the same
// wire backs it. Endpoints: GET / (the page), POST /chat, GET /memory, GET /health.
package server

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
)

// Deps wires the server to a live agent + the Cortex store it reads/writes.
type Deps struct {
	Agent *agent.Agent
	Store *cortex.MemoryStore
	Addr  string
}

// Serve runs the HTTP server until it errors.
func Serve(d Deps) error {
	addr := d.Addr
	if addr == "" {
		addr = ":8420"
	}
	return http.ListenAndServe(addr, Handler(d.Agent, d.Store))
}

// Handler builds the HTTP handler (exposed for tests).
func Handler(ag *agent.Agent, st *cortex.MemoryStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
		ns := st.All()
		writeJSON(w, map[string]any{"count": len(ns), "neurons": ns})
	})
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Prompt string `json:"prompt"`
			Mode   string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := ag.RunFull(r.Context(), strings.TrimSpace(req.Prompt))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{
			"reply":    res.Reply,
			"trace":    res.Trace,
			"recalled": res.Recalled,
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(uiHTML)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// uiHTML is the polished book-spine UI (E70 / F53): vertical spines, panels that
// open/close by clicking a spine, two side by side, drag & drop reorder - wired
// live to /chat, /memory, /health. Embedded so the binary stays single and
// dependency-free (no build step, no CDN).
//
//go:embed ui.html
var uiHTML []byte
