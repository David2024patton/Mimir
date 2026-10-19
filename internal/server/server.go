package server

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/engines"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/relay"
	"github.com/David2024patton/Mimir/internal/swarm"
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
		addr = "127.0.0.1:8420"
	}
	return http.ListenAndServe(addr, Handler(d.Agent, d.Store, d.Sessions, d.Auth, d.Usage))
}

func Handler(ag *agent.Agent, st cortex.Store, sessions cortex.SessionStore, auth *cortex.AuthStore, usage *llm.Tracker) http.Handler {
	mux := http.NewServeMux()
	relayMgr := relay.NewManager()

	// Per-session permission modes (N2): a mode set for one session must never
	// mutate the shared agent's mode for every other user. /chat resolves the
	// effective mode per request: an explicit request mode wins, then the
	// session's stored mode, then the agent default.
	var modeMu sync.Mutex
	sessionModes := map[string]string{}
	setSessionMode := func(sessionID, mode string) {
		modeMu.Lock()
		defer modeMu.Unlock()
		sessionModes[sessionID] = mode
	}
	sessionMode := func(sessionID string) string {
		modeMu.Lock()
		defer modeMu.Unlock()
		return sessionModes[sessionID]
	}
	requestMode := func(explicit, sessionID string) string {
		if m, ok := parseMode(explicit); ok {
			return m
		}
		if sessionID != "" {
			return sessionMode(sessionID)
		}
		return ""
	}

	// Public
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})

	// Public E2EE Pair & Mobile endpoints
	mux.HandleFunc("/api/pair", func(w http.ResponseWriter, r *http.Request) {
		sess, err := relayMgr.CreateSession(r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		asciiQR, _ := sess.GenerateQRCodeASCII()
		writeJSON(w, map[string]any{
			"session":  sess,
			"pair_uri": sess.PairURI(),
			"ascii_qr": asciiQR,
		})
	})
	mux.HandleFunc("/api/pair/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionID string `json:"session_id"`
			ClientPub string `json:"client_pubkey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sess, ok := relayMgr.GetSession(req.SessionID)
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		if err := sess.CompletePairing(req.ClientPub); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"status": "paired", "session_id": sess.ID})
	})
	mux.HandleFunc("/api/mobile/stream", func(w http.ResponseWriter, r *http.Request) {
		sessID := r.URL.Query().Get("id")
		sess, ok := relayMgr.GetSession(sessID)
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		sess.StreamHandler()(w, r)
	})
	mux.HandleFunc("/api/mobile/steer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionID string `json:"session_id"`
			Message   string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ag.Steer(req.Message)
		writeJSON(w, map[string]any{"status": "steered", "message": req.Message})
	})
	mux.HandleFunc("/api/mobile/mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionID string `json:"session_id"`
			Mode      string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mode, ok := parseMode(req.Mode)
		if !ok {
			http.Error(w, "invalid mode: ro, edit, or bypass required", http.StatusBadRequest)
			return
		}
		if req.SessionID == "" {
			// A mode without a session would have to mutate the shared agent,
			// which is exactly the N2 leak; modes are per-session state now.
			http.Error(w, "session_id is required: modes are stored per session", http.StatusBadRequest)
			return
		}
		setSessionMode(req.SessionID, mode)
		writeJSON(w, map[string]any{"status": "ok", "mode": mode, "session_id": req.SessionID})
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		pid, model, local := ag.ModelInfo()
		sdlcMgr := ag.SDLC()
		currPhase := ""
		if sdlcMgr != nil {
			currPhase = sdlcMgr.CurrentPhase().Title
		}
		writeJSON(w, map[string]any{
			"status":      "ok",
			"provider":    pid,
			"model":       model,
			"local":       local,
			"mode":        ag.Mode(),
			"effort":      ag.Effort(),
			"smol":        ag.SmolMode(),
			"sdlc_phase":  currPhase,
			"attribution": engines.AttributionNotice,
		})
	})
	mux.HandleFunc("/auth/send-code", sendCodeHandler(auth))
	mux.HandleFunc("/auth/verify-code", verifyCodeHandler(auth))

	// Landing page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(landingHTML)
	})

	// Auth page
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(authHTML)
	})

	// App page (protected client-side via token redirect)
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
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
		callerID := userIDFromContext(r.Context())
		sess, err := sessions.GetSession(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// IDOR ownership check (F12): user can only access their own sessions unless local mode
		if callerID != "" && callerID != "local" && sess.UserID != "" && sess.UserID != callerID {
			http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
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
		neurons := st.All()
		writeJSON(w, map[string]any{"count": len(neurons), "neurons": neurons})
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
			callerID := userIDFromContext(r.Context())
			sess, err := sessions.GetSession(r.Context(), req.SessionID)
			if err != nil {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			if callerID != "" && callerID != "local" && sess.UserID != "" && sess.UserID != callerID {
				http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
				return
			}
			if len(sess.Messages) > 0 {
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
		res, err := ag.RunFullWithMode(r.Context(), requestMode(req.Mode, req.SessionID), strings.TrimSpace(req.Prompt), history)
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
			callerID := userIDFromContext(r.Context())
			sess, err := sessions.GetSession(r.Context(), req.SessionID)
			if err != nil {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			if callerID != "" && callerID != "local" && sess.UserID != "" && sess.UserID != callerID {
				http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
				return
			}
			if len(sess.Messages) > 0 {
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
		err := ag.RunStreamWithMode(r.Context(), requestMode(req.Mode, req.SessionID), strings.TrimSpace(req.Prompt), history, func(e agent.Event) {
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

	api.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		pid, model, local := ag.ModelInfo()
		sdlcMgr := ag.SDLC()
		currPhase := ""
		if sdlcMgr != nil {
			currPhase = sdlcMgr.CurrentPhase().Title
		}
		writeJSON(w, map[string]any{
			"status":      "ok",
			"provider":    pid,
			"model":       model,
			"local":       local,
			"mode":        ag.Mode(),
			"effort":      ag.Effort(),
			"smol":        ag.SmolMode(),
			"sdlc_phase":  currPhase,
			"attribution": engines.AttributionNotice,
		})
	})
	api.HandleFunc("/engines", func(w http.ResponseWriter, r *http.Request) {
		m, err := engines.LoadManifest(".")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, m)
	})
	api.HandleFunc("/swarm", func(w http.ResponseWriter, r *http.Request) {
		f := swarm.NewFactory(swarm.FactoryConfig{
			RootDir:      ".",
			MaxWorkers:   4,
			TargetBranch: "main",
		})
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, map[string]any{
				"tasks": f.ListTasks(),
			})
		case http.MethodPost:
			var req struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			t, err := f.Submit(req.Title, req.Description)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, t)
		default:
			http.Error(w, "GET or POST only", http.StatusMethodNotAllowed)
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
	mux.Handle("/engines", authMiddleware(auth)(api))
	mux.Handle("/swarm", authMiddleware(auth)(api))
	mux.Handle("/api/engines", authMiddleware(auth)(api))
	mux.Handle("/api/swarm", authMiddleware(auth)(api))

	return mux
}

// parseMode validates a client-supplied permission mode string and returns
// its canonical form.
func parseMode(s string) (string, bool) {
	switch m := strings.ToLower(strings.TrimSpace(s)); m {
	case agent.ModeReadOnly, agent.ModeEdit, agent.ModeBypass:
		return m, true
	}
	return "", false
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

//go:embed ui.html
var uiHTML []byte

//go:embed landing.html
var landingHTML []byte

//go:embed auth.html
var authHTML []byte
