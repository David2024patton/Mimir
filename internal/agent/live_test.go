package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TestLiveRun exercises the full recall -> infer -> remember loop against a REAL
// provider (not a fake server). It is SKIPPED unless MIMIR_LIVE_BASE_URL is set, so
// the normal `go test ./...` stays green and offline. This is the framework's standing
// rule: after every build / feature, run a live test:
//
//	MIMIR_LIVE_BASE_URL=http://localhost:11434/v1 MIMIR_LIVE_MODEL=qwen2.5:0.5b \
//	  go test ./internal/agent -run TestLiveRun -v
//
// Add MIMIR_LIVE_API_KEY for cloud endpoints. The key is read from the environment and
// never printed.
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

	reply, err := a.Run(context.Background(), "Say hello in one short word.")
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
