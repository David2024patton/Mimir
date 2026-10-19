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

// TestPerRequestModeDoesNotLeak proves the N2 fix at the agent level: a run
// under a bypass override executes with bypass semantics, yet the agent's
// stored mode is untouched, so the next default run still runs read-only.
func TestPerRequestModeDoesNotLeak(t *testing.T) {
	dir := t.TempDir()
	call := llm.ToolCall{ID: "w1", Name: "write_file", Arguments: `{"path":"out.txt","content":"hi"}`}
	a := New(Config{
		Provider: &oneCallProvider{call: call}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeReadOnly,
	})

	res, err := a.RunFullWithMode(context.Background(), ModeBypass, "write it", nil)
	if err != nil {
		t.Fatalf("RunFullWithMode: %v", err)
	}
	if len(res.Trace) != 1 || res.Trace[0].Err != "" {
		t.Fatalf("bypass run: expected the write to succeed, trace = %+v", res.Trace)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.txt")); err != nil {
		t.Errorf("bypass run: file was not written: %v", err)
	}
	if got := a.Mode(); got != ModeReadOnly {
		t.Fatalf("agent mode leaked: stored mode = %q, want %q", got, ModeReadOnly)
	}

	// A subsequent default run must still be read-only: the write is blocked.
	a2 := New(Config{
		Provider: &oneCallProvider{call: call}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeReadOnly,
	})
	res, err = a2.RunFull(context.Background(), "write it", nil)
	if err != nil {
		t.Fatalf("RunFull: %v", err)
	}
	if len(res.Trace) != 1 || !strings.Contains(res.Trace[0].Err, "read-only") {
		t.Errorf("default run after bypass: expected a read-only block, trace = %+v", res.Trace)
	}
}
