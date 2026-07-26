// Package server is Mímir's live HTTP interface: it lets a browser (or any client)
// talk to the brain. GET / serves the polished book-spine UI (E70 / F53); the same
// wire backs it. Endpoints: GET / (the page), POST /chat, POST /chat/stream (SSE,
// F2.1), GET /memory, GET /usage (E69), GET /health, POST /sessions (F2.2).
package server

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
)

// Deps wires the server to a live agent + the Cortex store it reads/writes.
type Deps struct {
	Agent    *agent.Agent
	Store    cortex.Store
	Sessions *cortex.SessionStore // F2.2: conversation history
	Usage    *llm.Tracker
	Addr     string
}

// Serve runs the HTTP server until it errors.
func Serve(d Deps) error {
	addr := d.Addr
	if addr == "" {
		addr = ":8420"
	}
	return http.ListenAndServe(addr, Handler(d.Agent, d.Store, d.Sessions, d.Usage))
}

// Handler builds the HTTP handler (exposed for tests).
func Handler(ag *agent.Agent, st cortex.Store, sessions *cortex.SessionStore, usage *llm.Tracker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if sessions == nil {
				writeJSON(w, map[string]any{"sessions": []any{}})
				return
			}
			list, err := sessions.ListSessions(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{"sessions": list})
		case http.MethodPost:
			if sessions == nil {
				http.Error(w, "sessions not available", http.StatusServiceUnavailable)
				return
			}
			var req struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.Title == "" {
				req.Title = "New Conversation"
			}
			sess, err := sessions.CreateSession(r.Context(), req.Title)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, sess)
		default:
			http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if sessions == nil {
			http.Error(w, "sessions not available", http.StatusServiceUnavailable)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/sessions/")
		if id == "" {
			http.Error(w, "session id required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sess, err := sessions.GetSession(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, sess)
		case http.MethodDelete:
			if err := sessions.DeleteSession(r.Context(), id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]string{"status": "deleted"})
		case http.MethodPut:
			var req struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := sessions.UpdateTitle(r.Context(), id, req.Title); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]string{"status": "updated"})
		default:
			http.Error(w, "GET, PUT, or DELETE only", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
		ns := st.All()
		writeJSON(w, map[string]any{"count": len(ns), "neurons": ns})
	})
	mux.HandleFunc("/usage", func(w http.ResponseWriter, r *http.Request) {
		models := []llm.ModelUsage{}
		total := 0
		if usage != nil {
			if snap := usage.Snapshot(); snap != nil {
				models = snap
			}
			total = usage.TotalTokens()
		}
		writeJSON(w, map[string]any{"total_tokens": total, "models": models})
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
	mux.HandleFunc("/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Prompt    string `json:"prompt"`
			Mode      string `json:"mode"`
			SessionID string `json:"session_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		
		// Save user message to session if session_id provided
		if sessions != nil && req.SessionID != "" {
			_ = sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role:    "user",
				Content: req.Prompt,
			})
		}
		
		var assistantReply string
		err := ag.RunStream(r.Context(), strings.TrimSpace(req.Prompt), func(e agent.Event) {
			b, err := json.Marshal(e)
			if err != nil {
				return
			}
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
			if e.Type == agent.EventDone {
				assistantReply = e.Reply
			}
		})
		
		// Save assistant reply to session if session_id provided
		if sessions != nil && req.SessionID != "" && assistantReply != "" {
			_ = sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role:    "assistant",
				Content: assistantReply,
			})
		}
		
		if err != nil {
			_, _ = w.Write([]byte("data: " + `{"type":"error","err":"` + err.Error() + `"}` + "\n\n"))
			flusher.Flush()
		}
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
