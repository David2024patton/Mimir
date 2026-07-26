package server

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
)

type Deps struct {
	Agent    *agent.Agent
	Store    cortex.Store
	Sessions cortex.SessionStore
	Auth     *cortex.AuthStore
	Usage    *llm.Tracker
	Addr     string
}

func Serve(d Deps) error {
	addr := d.Addr
	if addr == "" {
		addr = ":8420"
	}
	return http.ListenAndServe(addr, Handler(d.Agent, d.Store, d.Sessions, d.Auth, d.Usage))
}

func Handler(ag *agent.Agent, st cortex.Store, sessions cortex.SessionStore, auth *cortex.AuthStore, usage *llm.Tracker) http.Handler {
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/auth/send-code", sendCodeHandler(auth))
	mux.HandleFunc("/auth/verify-code", verifyCodeHandler(auth))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(uiHTML)
	})

	// Protected API routes (built as a sub-mux then wrapped with auth)
	api := http.NewServeMux()
	api.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		userID := userIDFromContext(r.Context())
		switch r.Method {
		case http.MethodGet:
			list, err := sessions.ListSessions(r.Context(), userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{"sessions": list})
		case http.MethodPost:
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
			sess, err := sessions.CreateSession(r.Context(), req.Title, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, sess)
		default:
			http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
		}
	})
	api.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
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
	api.HandleFunc("/auth/me", meHandler(auth))
	api.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"neurons": st.All()})
	})
	api.HandleFunc("/usage", func(w http.ResponseWriter, r *http.Request) {
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
	api.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
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
		var history []llm.Message
		if req.SessionID != "" {
			sess, err := sessions.GetSession(r.Context(), req.SessionID)
			if err == nil && len(sess.Messages) > 0 {
				for _, m := range sess.Messages {
					history = append(history, llm.Message{Role: m.Role, Content: m.Content})
				}
			}
			if err := sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role: "user", Content: req.Prompt,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res, err := ag.RunFull(r.Context(), strings.TrimSpace(req.Prompt), history)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if req.SessionID != "" && res.Reply != "" {
			if err := sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role: "assistant", Content: res.Reply,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		writeJSON(w, map[string]any{
			"reply":    res.Reply,
			"trace":    res.Trace,
			"recalled": res.Recalled,
		})
	})
	api.HandleFunc("/chat/stream", func(w http.ResponseWriter, r *http.Request) {
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
		var history []llm.Message
		if req.SessionID != "" {
			sess, err := sessions.GetSession(r.Context(), req.SessionID)
			if err == nil && len(sess.Messages) > 0 {
				for _, m := range sess.Messages {
					history = append(history, llm.Message{Role: m.Role, Content: m.Content})
				}
			}
			_ = sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role:    "user",
				Content: req.Prompt,
			})
		}
		var assistantReply string
		err := ag.RunStream(r.Context(), strings.TrimSpace(req.Prompt), history, func(e agent.Event) {
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
		if req.SessionID != "" && assistantReply != "" {
			_ = sessions.AppendMessage(r.Context(), req.SessionID, cortex.Message{
				Role:    "assistant",
				Content: assistantReply,
			})
		}
		if err != nil {
			_, _ = w.Write([]byte(fmt.Sprintf("data: {\"type\":\"error\",\"err\":%q}\n\n", err.Error())))
			flusher.Flush()
		}
	})

	// Wrap the api mux with auth middleware and register on main mux
	mux.Handle("/sessions", authMiddleware(auth)(api))
	mux.Handle("/sessions/", authMiddleware(auth)(api))
	mux.Handle("/auth/me", authMiddleware(auth)(api))
	mux.Handle("/memory", authMiddleware(auth)(api))
	mux.Handle("/usage", authMiddleware(auth)(api))
	mux.Handle("/chat", authMiddleware(auth)(api))
	mux.Handle("/chat/stream", authMiddleware(auth)(api))

	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

//go:embed ui.html
var uiHTML []byte
