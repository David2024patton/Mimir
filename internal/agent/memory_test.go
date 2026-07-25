package agent

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// textProvider always returns a fixed reply with no tool calls - enough to drive the
// loop and exercise memory write/recall deterministically (no network).
type textProvider struct{ t string }

func (p *textProvider) ID() string { return "txt" }
func (p *textProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{Content: p.t, FinishReason: "stop"}, nil
}
func (p *textProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

// TestCrossRunMemory proves memory survives across runs: run 1 (a fresh agent + store)
// writes a memory to a file-backed Cortex; a second agent opening a NEW store from the
// same file (simulating a process restart) recalls it on run 2. No network involved.
func TestCrossRunMemory(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cortex.json")
	mk := func() *Agent {
		st, err := cortex.NewMemoryStoreAt(p)
		if err != nil {
			t.Fatalf("store: %v", err)
		}
		return New(Config{
			Provider: &textProvider{"ok"}, Tools: tools.NewRegistry(), Cortex: st, Model: "m",
		})
	}

	if _, err := mk().Run(context.Background(), "remember the codeword KIWI55"); err != nil {
		t.Fatalf("run1: %v", err)
	}
	res, err := mk().RunFull(context.Background(), "tell me about KIWI55")
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
		t.Fatalf("expected cross-run recall of KIWI55; recalled=%v", res.Recalled)
	}
}
