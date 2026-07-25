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
