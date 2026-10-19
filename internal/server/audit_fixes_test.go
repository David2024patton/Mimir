package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// steerProbeProvider drives one tool call on its first Generate so the agent
// loop reaches the steer-drain point, and records every message it is shown
// so tests can prove whether a steering instruction reached the model.
type steerProbeProvider struct {
	mu    sync.Mutex
	calls int
	seen  []string
}

func (p *steerProbeProvider) ID() string { return "steerprobe" }
func (p *steerProbeProvider) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	for _, m := range req.Messages {
		p.seen = append(p.seen, m.Content)
	}
	if p.calls == 1 {
		return llm.GenerateResponse{
			ToolCalls:    []llm.ToolCall{{ID: "tc1", Name: "todo", Arguments: "{}"}},
			FinishReason: "tool_calls",
		}, nil
	}
	return llm.GenerateResponse{Content: "probe reply", FinishReason: "stop"}, nil
}
func (p *steerProbeProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}
func (p *steerProbeProvider) sawSteering(msg string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range p.seen {
		if strings.Contains(s, "[User steering instruction]: "+msg) {
			return true
		}
	}
	return false
}

// mobileFixture builds a Handler with a real session store holding one
// session owned by user_alice and one by user_bob, plus the probe agent.
type mobileFixture struct {
	handler http.Handler
	prov    *steerProbeProvider
	alice   string // alice's session ID
	bob     string // bob's session ID
}

func newMobileFixture(t *testing.T) *mobileFixture {
	t.Helper()
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	sessStore, err := cortex.NewFileSessionStore(t.TempDir())
	if err != nil {
		t.Fatalf("session store: %v", err)
	}
	sessA, err := sessStore.CreateSession(context.Background(), "Alice chat", "user_alice")
	if err != nil {
		t.Fatalf("create alice session: %v", err)
	}
	sessB, err := sessStore.CreateSession(context.Background(), "Bob chat", "user_bob")
	if err != nil {
		t.Fatalf("create bob session: %v", err)
	}
	prov := &steerProbeProvider{}
	ag := agent.New(agent.Config{Provider: prov, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	return &mobileFixture{
		handler: Handler(ag, store, sessStore, nil, nil),
		prov:    prov,
		alice:   sessA.ID,
		bob:     sessB.ID,
	}
}

// postAs issues a POST with the caller identity injected into the request
// context, following the fake-auth pattern of TestServerSessionIDORProtection
// (the auth middleware passes a pre-set identity through in local mode).
func postAs(h http.Handler, path, body, userID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), ctxUserID, userID))
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func (f *mobileFixture) chat(t *testing.T) {
	t.Helper()
	w := postAs(f.handler, "/chat", `{"prompt":"go"}`, "")
	if w.Code != http.StatusOK {
		t.Fatalf("probe /chat status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
}

// TestMobileEndpointsRequireAuthWhenConfigured proves FIX 1's first half:
// with an AuthStore configured, /api/mobile/steer and /api/mobile/mode reject
// requests carrying no credentials, like /chat does. (On the unfixed code
// both are registered on the public mux and answer 200.)
func TestMobileEndpointsRequireAuthWhenConfigured(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	ag := agent.New(agent.Config{Provider: &fakeText{"ok"}, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	// A non-nil AuthStore switches the middleware to credential checking.
	// No Authorization header is sent, so validation is never reached and
	// the nil client is never dereferenced.
	auth := cortex.NewAuthStore(nil)
	srv := httptest.NewServer(Handler(ag, store, nil, auth, nil))
	defer srv.Close()

	for _, path := range []string{"/api/mobile/steer", "/api/mobile/mode"} {
		resp, err := http.Post(srv.URL+path, "application/json",
			strings.NewReader(`{"session_id":"s1","message":"hi","mode":"bypass"}`))
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s without credentials: status = %d, want 401", path, resp.StatusCode)
		}
	}
}

// TestMobileSteerForeignSessionRejected proves a caller cannot steer another
// user's session, and that the rejected steer never enters the agent's queue.
func TestMobileSteerForeignSessionRejected(t *testing.T) {
	f := newMobileFixture(t)
	w := postAs(f.handler, "/api/mobile/steer",
		`{"session_id":"`+f.alice+`","message":"pwned steer"}`, "user_bob")
	if w.Code != http.StatusForbidden {
		t.Fatalf("foreign steer status = %d, want 403 (body: %s)", w.Code, w.Body.String())
	}
	f.chat(t)
	if f.prov.sawSteering("pwned steer") {
		t.Error("rejected foreign steer reached the agent run: steer queue was polluted")
	}
}

// TestMobileSteerUnknownSessionRejected proves a steer naming a session that
// does not exist is rejected and not enqueued.
func TestMobileSteerUnknownSessionRejected(t *testing.T) {
	f := newMobileFixture(t)
	w := postAs(f.handler, "/api/mobile/steer",
		`{"session_id":"no-such-session","message":"ghost steer"}`, "user_alice")
	if w.Code != http.StatusForbidden {
		t.Fatalf("unknown-session steer status = %d, want 403", w.Code)
	}
	f.chat(t)
	if f.prov.sawSteering("ghost steer") {
		t.Error("unknown-session steer reached the agent run")
	}
}

// TestMobileSteerOwnSessionAccepted is the positive control: the owner can
// steer their own session and the instruction reaches the next run OF THAT
// SESSION (delivery is session-bound, so the run must carry the session ID).
func TestMobileSteerOwnSessionAccepted(t *testing.T) {
	f := newMobileFixture(t)
	w := postAs(f.handler, "/api/mobile/steer",
		`{"session_id":"`+f.alice+`","message":"please pivot"}`, "user_alice")
	if w.Code != http.StatusOK {
		t.Fatalf("own steer status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	w = postAs(f.handler, "/chat",
		`{"prompt":"go","session_id":"`+f.alice+`"}`, "user_alice")
	if w.Code != http.StatusOK {
		t.Fatalf("own-session /chat status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if !f.prov.sawSteering("please pivot") {
		t.Error("accepted own-session steer never reached the agent run")
	}
}

// TestMobileSteerLocalModeStillWorks proves the no-auth local setup (no
// session store wired) keeps accepting steers as before.
func TestMobileSteerLocalModeStillWorks(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	ag := agent.New(agent.Config{Provider: &fakeText{"ok"}, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	srv := httptest.NewServer(Handler(ag, store, nil, nil, nil))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/mobile/steer", "application/json",
		strings.NewReader(`{"session_id":"anything","message":"local steer"}`))
	if err != nil {
		t.Fatalf("post steer: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("local steer status = %d, want 200", resp.StatusCode)
	}
}

// TestMobileModeOwnership proves the mode endpoint applies the same
// ownership rule: foreign and unknown sessions are rejected, the owner's
// own session accepts the mode (stored per session, as before).
func TestMobileModeOwnership(t *testing.T) {
	f := newMobileFixture(t)

	w := postAs(f.handler, "/api/mobile/mode",
		`{"session_id":"`+f.alice+`","mode":"bypass"}`, "user_bob")
	if w.Code != http.StatusForbidden {
		t.Errorf("foreign mode set status = %d, want 403", w.Code)
	}

	w = postAs(f.handler, "/api/mobile/mode",
		`{"session_id":"no-such-session","mode":"bypass"}`, "user_alice")
	if w.Code != http.StatusForbidden {
		t.Errorf("unknown-session mode set status = %d, want 403", w.Code)
	}

	w = postAs(f.handler, "/api/mobile/mode",
		`{"session_id":"`+f.alice+`","mode":"bypass"}`, "user_alice")
	if w.Code != http.StatusOK {
		t.Errorf("own mode set status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	// Local mode (caller "local") may set a mode on any existing session,
	// matching the /chat and item-route ownership rule.
	w = postAs(f.handler, "/api/mobile/mode",
		`{"session_id":"`+f.bob+`","mode":"ro"}`, "")
	if w.Code != http.StatusOK {
		t.Errorf("local mode set status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
}

// errStreamProvider fails the stream immediately with an error whose text
// contains control characters (FIX 3 probe).
type errStreamProvider struct{ err error }

func (p *errStreamProvider) ID() string { return "errstream" }
func (p *errStreamProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{}, p.err
}
func (p *errStreamProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, p.err
}

// TestChatStreamErrorFrameIsValidJSON proves FIX 3: the /chat/stream error
// frame is marshaled JSON, so an error containing control bytes still
// produces a parseable frame carrying the exact message. (The hand-rolled
// %q frame emitted \xNN escapes, which are not valid JSON.)
func TestChatStreamErrorFrameIsValidJSON(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	wantErr := "stream blew up\x01\x02badly"
	ag := agent.New(agent.Config{
		Provider: &errStreamProvider{err: errors.New(wantErr)},
		Tools:    tools.NewRegistry(), Cortex: store, Model: "m",
	})
	srv := httptest.NewServer(Handler(ag, store, nil, nil, nil))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/chat/stream", "application/json", strings.NewReader(`{"prompt":"hi"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	body := new(strings.Builder)
	buf := make([]byte, 4096)
	for {
		n, rerr := resp.Body.Read(buf)
		body.Write(buf[:n])
		if rerr != nil {
			break
		}
	}
	var last string
	for _, line := range strings.Split(body.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			last = strings.TrimPrefix(line, "data: ")
		}
	}
	if last == "" {
		t.Fatalf("no data frames in stream response: %q", body.String())
	}
	var frame map[string]any
	if err := json.Unmarshal([]byte(last), &frame); err != nil {
		t.Fatalf("final error frame is not valid JSON: %v (frame: %s)", err, last)
	}
	if frame["type"] != "error" {
		t.Errorf("frame type = %v, want error", frame["type"])
	}
	if frame["err"] != wantErr {
		t.Errorf("frame err = %q, want the exact provider error %q", frame["err"], wantErr)
	}
}
