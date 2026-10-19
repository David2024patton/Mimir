package server

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// steerRunProvider is a run-aware recording provider: the first Generate of
// each run (messages end with the fresh user prompt) returns one tool call so
// the loop reaches the steer-drain point; the follow-up Generate (messages
// end with the tool result) returns final text. It records every message it
// is shown so tests can prove which run a steering instruction reached.
type steerRunProvider struct {
	mu   sync.Mutex
	seen []string
}

func (p *steerRunProvider) ID() string { return "steerrun" }
func (p *steerRunProvider) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, m := range req.Messages {
		p.seen = append(p.seen, m.Content)
	}
	if n := len(req.Messages); n == 0 || req.Messages[n-1].Role != "tool" {
		return llm.GenerateResponse{
			ToolCalls:    []llm.ToolCall{{ID: "tc1", Name: "todo", Arguments: "{}"}},
			FinishReason: "tool_calls",
		}, nil
	}
	return llm.GenerateResponse{Content: "run reply", FinishReason: "stop"}, nil
}
func (p *steerRunProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}
func (p *steerRunProvider) sawSteering(msg string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range p.seen {
		if strings.Contains(s, "[User steering instruction]: "+msg) {
			return true
		}
	}
	return false
}

// steerSessionFixture is newMobileFixture with the run-aware provider, so a
// test can steer one session and then run /chat on a specific other session.
type steerSessionFixture struct {
	handler http.Handler
	prov    *steerRunProvider
	alice   string
	bob     string
}

func newSteerSessionFixture(t *testing.T) *steerSessionFixture {
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
	prov := &steerRunProvider{}
	ag := agent.New(agent.Config{Provider: prov, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	return &steerSessionFixture{
		handler: Handler(ag, store, sessStore, nil, nil),
		prov:    prov,
		alice:   sessA.ID,
		bob:     sessB.ID,
	}
}

func (f *steerSessionFixture) chatAs(t *testing.T, userID, sessionID string) {
	t.Helper()
	body := `{"prompt":"go"}`
	if sessionID != "" {
		body = `{"prompt":"go","session_id":"` + sessionID + `"}`
	}
	w := postAs(f.handler, "/chat", body, userID)
	if w.Code != http.StatusOK {
		t.Fatalf("chat as %q on session %q: status = %d, want 200 (body: %s)",
			userID, sessionID, w.Code, w.Body.String())
	}
}

// TestMobileSteerStaysInOwnSession is the FIX 1 cross-user delivery proof:
// Bob steers HIS OWN session (accepted, as before), but the instruction may
// only ever be delivered into a run of Bob's session. Alice's run on her own
// session and a sessionless run must never see it; Bob's own run must.
// (On the pre-fix code the steer sat in one global channel and Alice's run,
// whichever came first, drained it.)
func TestMobileSteerStaysInOwnSession(t *testing.T) {
	f := newSteerSessionFixture(t)

	w := postAs(f.handler, "/api/mobile/steer",
		`{"session_id":"`+f.bob+`","message":"bob pivot now"}`, "user_bob")
	if w.Code != http.StatusOK {
		t.Fatalf("bob steer own session: status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	// Alice runs /chat on HER session: Bob's steer must not reach her model.
	f.chatAs(t, "user_alice", f.alice)
	if f.prov.sawSteering("bob pivot now") {
		t.Fatal("cross-user leak: Bob's steer was delivered into Alice's run")
	}

	// A sessionless run drains nothing either (fail closed).
	f.chatAs(t, "", "")
	if f.prov.sawSteering("bob pivot now") {
		t.Fatal("sessionless run received a session-bound steer")
	}

	// Bob runs /chat on HIS session: his steer is delivered there.
	f.chatAs(t, "user_bob", f.bob)
	if !f.prov.sawSteering("bob pivot now") {
		t.Fatal("Bob's steer was never delivered into Bob's own run")
	}
}
