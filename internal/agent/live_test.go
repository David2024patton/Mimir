package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TestLiveRun exercises the recall -> infer -> remember loop against a REAL provider
// (not a fake server). SKIPPED unless MIMIR_LIVE_BASE_URL is set, so the normal
// `go test ./...` stays green and offline. See docs/TESTING-POLICY.md.
func TestLiveRun(t *testing.T) {
	base := os.Getenv("MIMIR_LIVE_BASE_URL")
	if base == "" {
		t.Skip("set MIMIR_LIVE_BASE_URL (+ MIMIR_LIVE_MODEL, optional MIMIR_LIVE_API_KEY) to run the live test")
	}
	model := os.Getenv("MIMIR_LIVE_MODEL")
	if model == "" {
		model = "qwen2.5:0.5b"
	}

	store := cortex.NewMemoryStore()
	if _, err := store.PutNeuron(context.Background(), cortex.Neuron{
		Kind: cortex.KindMemory, Layer: "preference", Content: "the secret codeword is pinecone", Decay: 1,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	a := New(Config{
		Provider: &llm.OpenAIProvider{
			IDStr: "live", BaseURL: base, APIKey: os.Getenv("MIMIR_LIVE_API_KEY"), Model: model,
		},
		Tools: tools.NewRegistry(), Cortex: store, Model: model,
	})

	reply, err := a.Run(context.Background(), "Say hello in one short word.", nil)
	if err != nil {
		t.Fatalf("live Run: %v", err)
	}
	if strings.TrimSpace(reply) == "" {
		t.Fatal("live model returned an empty reply")
	}
	t.Logf("live reply from %s/%s: %q", "live", model, reply)

	remembered, _ := store.Search(context.Background(), "hello", 5)
	if len(remembered) == 0 {
		t.Error("expected the live exchange to be remembered in the Cortex")
	}
}

// TestLiveToolCall is the live proof that a real model can drive a real tool through
// the loop. SKIPPED unless MIMIR_LIVE_TOOL=1 (it needs a tool-capable model, e.g.
// qwen3:8b). It asks the model to run bash and checks the tool actually executed.
func TestLiveToolCall(t *testing.T) {
	if os.Getenv("MIMIR_LIVE_TOOL") == "" {
		t.Skip("set MIMIR_LIVE_TOOL=1 (and MIMIR_LIVE_BASE_URL + a tool-capable MIMIR_LIVE_MODEL like qwen3:8b) to run the live tool test")
	}
	base := os.Getenv("MIMIR_LIVE_BASE_URL")
	if base == "" {
		t.Skip("MIMIR_LIVE_TOOL=1 but MIMIR_LIVE_BASE_URL is unset")
	}
	model := os.Getenv("MIMIR_LIVE_MODEL")
	if model == "" {
		model = "qwen3:8b"
	}

	a := New(Config{
		Provider: &llm.OpenAIProvider{
			IDStr: "live", BaseURL: base, APIKey: os.Getenv("MIMIR_LIVE_API_KEY"), Model: model,
		},
		Tools: tools.Default(t.TempDir()), Cortex: cortex.NewMemoryStore(), Model: model,
	})

	reply, trace, err := a.RunTrace(context.Background(), "You have a bash tool. Use the bash tool to run exactly this command: echo MIMIR_TOOL_LIVE. Then reply with the single word DONE.", nil)
	if err != nil {
		t.Fatalf("live RunTrace: %v", err)
	}
	used := false
	for _, s := range trace {
		if s.Name == "bash" {
			used = true
			if !strings.Contains(s.Result, "MIMIR_TOOL_LIVE") {
				t.Errorf("bash result = %q, want it to contain MIMIR_TOOL_LIVE", s.Result)
			}
		}
	}
	if !used {
		t.Fatalf("the live model did not call the bash tool (reply was %q); pick a tool-capable model via MIMIR_LIVE_MODEL", reply)
	}
	t.Logf("live tool loop OK; reply=%q trace steps=%d", reply, len(trace))
}

// TestLiveCrossRunMemory proves cross-run memory with a REAL model: run 1 writes a
// memory to a file-backed Cortex; a fresh store reopened from the same file recalls it
// on run 2. Asserts on Recalled (deterministic regardless of the model's wording).
func TestLiveCrossRunMemory(t *testing.T) {
	if os.Getenv("MIMIR_LIVE_MEMORY") == "" {
		t.Skip("set MIMIR_LIVE_MEMORY=1 (+ MIMIR_LIVE_BASE_URL + model) to run the live cross-run memory test")
	}
	base := os.Getenv("MIMIR_LIVE_BASE_URL")
	if base == "" {
		t.Skip("MIMIR_LIVE_MEMORY=1 but MIMIR_LIVE_BASE_URL is unset")
	}
	model := os.Getenv("MIMIR_LIVE_MODEL")
	if model == "" {
		model = "qwen3:8b"
	}
	p := filepath.Join(t.TempDir(), "cortex.json")
	mk := func() *Agent {
		st, err := cortex.NewMemoryStoreAt(p)
		if err != nil {
			t.Fatalf("store: %v", err)
		}
		return New(Config{
			Provider: &llm.OpenAIProvider{
				IDStr: "live", BaseURL: base, APIKey: os.Getenv("MIMIR_LIVE_API_KEY"), Model: model,
			},
			Tools: tools.NewRegistry(), Cortex: st, Model: model,
		})
	}

	if _, err := mk().Run(context.Background(), "Remember this exactly: the launch code is KIWI55. Reply OK.", nil); err != nil {
		t.Fatalf("run1: %v", err)
	}
	res, err := mk().RunFull(context.Background(), "What is the launch code KIWI55? Answer from memory.", nil)
	if err != nil {
		t.Fatalf("run2: %v", err)
	}
	found := false
	for _, c := range res.Recalled {
		if strings.Contains(c, "KIWI55") {
			found = true
		}
	}
	if !found {
		t.Fatalf("run 2 did not recall run 1's memory from disk; recalled=%v", res.Recalled)
	}
	t.Logf("live cross-run memory OK; recalled %d memory(ies)", len(res.Recalled))
}
