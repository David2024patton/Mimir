package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

type fakeText struct{ t string }

func (f *fakeText) ID() string { return "fake" }
func (f *fakeText) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{Content: f.t, FinishReason: "stop"}, nil
}
func (f *fakeText) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

func TestServerChatAndMemory(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	ag := agent.New(agent.Config{
		Provider: &fakeText{"WIRED"}, Tools: tools.NewRegistry(), Cortex: store, Model: "m",
	})
	srv := httptest.NewServer(Handler(ag, store))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/chat", "application/json", strings.NewReader(`{"prompt":"hi","mode":"chat"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out struct {
		Reply string `json:"reply"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Reply != "WIRED" {
		t.Errorf("reply = %q", out.Reply)
	}

	mresp, err := http.Get(srv.URL + "/memory")
	if err != nil {
		t.Fatalf("memory: %v", err)
	}
	defer mresp.Body.Close()
	var mem struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(mresp.Body).Decode(&mem); err != nil {
		t.Fatalf("decode memory: %v", err)
	}
	if mem.Count < 1 {
		t.Errorf("memory count = %d, want >= 1 (the exchange should be persisted)", mem.Count)
	}

	hresp, err := http.Get(srv.URL + "/health")
	if err != nil || hresp.StatusCode != http.StatusOK {
		t.Fatalf("health: err=%v status=%v", err, hresp)
	}
}

// TestServerServesBookSpineUI proves GET / serves the polished book-spine UI (E70),
// not the old minimal shell: it must expose the spines + the live wire endpoints.
func TestServerServesBookSpineUI(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	ag := agent.New(agent.Config{Provider: &fakeText{"x"}, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	srv := httptest.NewServer(Handler(ag, store))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("get /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("content-type = %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	for _, want := range []string{"Mímir", "CORTEX", "spine", "/chat", "/memory", "/health"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("book-spine UI missing %q", want)
		}
	}
}

// TestServerMemoryJSONCasing proves neurons serialize with lowercase JSON keys
// (kind/content), which the book-spine UI reads. Without the struct tags this
// decodes empty and the Cortex panel would render blank.
func TestServerMemoryJSONCasing(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	if _, err := store.PutNeuron(context.Background(), cortex.Neuron{
		Kind: cortex.KindMemory, Content: "the launch code is KIWI55", Decay: 1,
	}); err != nil {
		t.Fatalf("put: %v", err)
	}
	ag := agent.New(agent.Config{Provider: &fakeText{"x"}, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	srv := httptest.NewServer(Handler(ag, store))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/memory")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	var out struct {
		Count   int `json:"count"`
		Neurons []struct {
			Kind    string `json:"kind"`
			Content string `json:"content"`
		} `json:"neurons"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Count != 1 {
		t.Fatalf("count = %d", out.Count)
	}
	if out.Neurons[0].Kind != "memory" || out.Neurons[0].Content != "the launch code is KIWI55" {
		t.Errorf("lowercase JSON keys not honored: %+v", out.Neurons[0])
	}
}

// fakeStream streams a final answer in two deltas (no tool calls) - enough to drive
// the SSE endpoint.
type fakeStream struct{}

func (f *fakeStream) ID() string { return "fakestream" }
func (f *fakeStream) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{Content: "WIRED", FinishReason: "stop"}, nil
}
func (f *fakeStream) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	out := make(chan llm.StreamEvent)
	go func() {
		defer close(out)
		out <- llm.StreamEvent{Delta: "WI"}
		out <- llm.StreamEvent{Delta: "RED"}
		out <- llm.StreamEvent{Done: true}
	}()
	return out, nil
}

// TestServerChatStream proves POST /chat/stream speaks SSE: token deltas arrive as
// separate events and a done event carries the full reply.
func TestServerChatStream(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	ag := agent.New(agent.Config{Provider: &fakeStream{}, Tools: tools.NewRegistry(), Cortex: store, Model: "m"})
	srv := httptest.NewServer(Handler(ag, store))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/chat/stream", "application/json", strings.NewReader(`{"prompt":"hi","mode":"chat"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("content-type = %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	var tokens, reply string
	var sawDone bool
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var e agent.Event
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &e); err != nil {
			continue
		}
		switch e.Type {
		case agent.EventToken:
			tokens += e.Text
		case agent.EventDone:
			sawDone = true
			reply = e.Reply
		case agent.EventError:
			t.Fatalf("unexpected error event: %s", e.Err)
		}
	}
	if tokens != "WIRED" {
		t.Errorf("streamed tokens = %q, want WIRED", tokens)
	}
	if !sawDone || reply != "WIRED" {
		t.Errorf("done reply = %q (sawDone=%v), want WIRED", reply, sawDone)
	}
}
