package server

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
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
	// F16 remainder: the default bind is loopback, but an explicitly configured
	// non-loopback address with no auth store (file backend) would expose an
	// unauthenticated server to the network. Refuse to start unless the
	// operator explicitly accepts that with MIMIR_ALLOW_UNAUTHENTICATED=1.
	if d.Auth == nil && os.Getenv("MIMIR_ALLOW_UNAUTHENTICATED") != "1" && !isLoopbackAddr(addr) {
		// Trust-boundary record (FIX 6): a refused unauthenticated bind is a
		// security event the operator should be able to find in the logs.
		slog.Default().Error("refusing unauthenticated non-loopback bind",
			"addr", addr, "opt_in_env", "MIMIR_ALLOW_UNAUTHENTICATED")
		return fmt.Errorf("refusing to serve on non-loopback address %s without authentication: use the surreal backend for auth, or set MIMIR_ALLOW_UNAUTHENTICATED=1 to accept the risk", addr)
	}
	return http.ListenAndServe(addr, Handler(d.Agent, d.Store, d.Sessions, d.Auth, d.Usage))
}

// isLoopbackAddr reports whether a listen address binds only the loopback
// interface. An empty host binds all interfaces and is not loopback.
func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
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

	// checkMobileSession enforces the F12 ownership rule for the mobile
	// endpoints (FIX 1): the target session must exist and belong to the
	// caller, using the same caller/owner test as the session item routes
	// and /chat. In local mode the caller is "local" and passes; with no
	// session store wired there is nothing to check against. On rejection
	// it writes the response plus a trust-boundary log record (FIX 6) and
	// returns false.
	checkMobileSession := func(w http.ResponseWriter, r *http.Request, action, sessionID string) bool {
		log := slog.Default()
		if sessions == nil {
			return true
		}
		sess, err := sessions.GetSession(r.Context(), sessionID)
		if err != nil {
			log.Warn("mobile request rejected: unknown session",
				"action", action, "session_id", sessionID, "caller", userIDFromContext(r.Context()))
			http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
			return false
		}
		callerID := userIDFromContext(r.Context())
		if callerID != "" && callerID != "local" && sess.UserID != "" && sess.UserID != callerID {
			log.Warn("mobile request rejected: foreign session",
				"action", action, "session_id", sessionID, "caller", callerID)
			http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
			return false
		}
		return true
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
			// PairingToken is the session nonce from the QR code / pair URI:
			// it proves the phone actually scanned the code, so a caller that
			// merely learned the session ID cannot hijack the pairing.
			PairingToken string `json:"pairing_token"`
			ClientPub    string `json:"client_pubkey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sess, ok := relayMgr.GetSession(req.SessionID)
		if !ok {
			slog.Default().Warn("pairing complete rejected: unknown session", "session_id", req.SessionID)
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		if err := sess.CompletePairing(req.PairingToken, req.ClientPub); err != nil {
			slog.Default().Warn("pairing complete failed", "session_id", req.SessionID, "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Default().Info("pairing complete", "session_id", sess.ID)
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
	// The mobile steer and mode endpoints sit behind the same auth
	// middleware as /chat (FIX 1): steer injects text into the agent's next
	// run and mode can arm bypass for a session, so neither may be public.
	// Each also binds to the caller's own session via checkMobileSession.
	mux.Handle("/api/mobile/steer", authMiddleware(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if req.SessionID == "" {
			slog.Default().Warn("mobile steer rejected: missing session", "caller", userIDFromContext(r.Context()))
			http.Error(w, "session_id is required: steers are bound to a session", http.StatusBadRequest)
			return
		}
		if !checkMobileSession(w, r, "steer", req.SessionID) {
			return
		}
		ag.Steer(req.Message)
		// FIX 6: log the acceptance with the message LENGTH only; steer
		// text is user content and never goes in the logs.
		slog.Default().Info("mobile steer accepted",
			"session_id", req.SessionID, "caller", userIDFromContext(r.Context()), "message_len", len(req.Message))
		writeJSON(w, map[string]any{"status": "steered", "message": req.Message})
	})))
	mux.Handle("/api/mobile/mode", authMiddleware(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if !checkMobileSession(w, r, "mode", req.SessionID) {
			return
		}
		setSessionMode(req.SessionID, mode)
		slog.Default().Info("mobile mode set",
			"session_id", req.SessionID, "caller", userIDFromContext(r.Context()), "mode", mode)
		writeJSON(w, map[string]any{"status": "ok", "mode": mode, "session_id": req.SessionID})
	})))
	// /api/status discloses provider, model, mode, and SDLC phase (N10): when
	// an auth store is configured it requires authentication like every other
	// API route; in local mode the middleware passes through (loopback-only
	// bind per the F16 fix), so local behavior is unchanged.
	mux.Handle("/api/status", authMiddleware(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})))
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
			// Marshal the error frame (FIX 3): the previous hand-rolled %q
			// frame emitted \xNN escapes for control bytes, which are not
			// valid JSON, so strict SSE consumers could not parse it.
			frame, _ := json.Marshal(map[string]string{"type": "error", "err": err.Error()})
			_, _ = w.Write([]byte("data: " + string(frame) + "\n\n"))
			flusher.Flush()
		}
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
