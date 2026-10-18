package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/llm"
)

func TestPruneHistoryPlaceholders(t *testing.T) {
	bigCode := strings.Repeat("func hello() {}\n", 40) // > 300 chars
	argsJSON := fmt.Sprintf(`{"path":"main.go","content":%q}`, bigCode)

	msgs := []llm.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "write main.go"},
		{
			Role: "assistant",
			ToolCalls: []llm.ToolCall{
				{ID: "tc1", Name: "write_file", Arguments: argsJSON},
			},
		},
		{Role: "tool", ToolCallID: "tc1", Content: "wrote main.go"},
		{Role: "user", Content: "now check status"},
		{Role: "assistant", Content: "checking"},
	}

	pruned := PruneHistoryPlaceholders(msgs, 2)
	gotArgs := pruned[2].ToolCalls[0].Arguments
	if !strings.Contains(gotArgs, "characters already applied to main.go") {
		t.Fatalf("expected placeholder in pruned write_file args, got: %s", gotArgs)
	}
	if len(gotArgs) >= len(argsJSON) {
		t.Fatalf("expected pruned args to be much shorter (%d vs %d)", len(gotArgs), len(argsJSON))
	}
}

func TestMaybeCompactContext(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	ag := New(Config{
		ContextBudget: 100, // very small budget so threshold is 80 tokens (~320 chars)
	})

	longContent := strings.Repeat("This is a long conversation content line to trigger token compaction. ", 20)
	msgs := []llm.Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Task 1 " + longContent},
		{Role: "assistant", Content: "Task 1 answer"},
		{Role: "user", Content: "Task 2"},
		{Role: "assistant", Content: "Task 2 answer"},
	}

	compacted := ag.maybeCompactContext(msgs, false)
	if len(compacted) >= len(msgs) {
		t.Fatalf("expected compacted messages to be trimmed, got %d msgs", len(compacted))
	}
	if !strings.Contains(compacted[1].Content, "Context Compaction") {
		t.Fatalf("expected compaction notice in message 1, got: %s", compacted[1].Content)
	}

	// Verify .mimir/HANDOFF.md was written
	handoffPath := filepath.Join(tmpDir, ".mimir", "HANDOFF.md")
	if _, err := os.Stat(handoffPath); os.IsNotExist(err) {
		t.Fatalf("expected %s to be created by auto-compaction", handoffPath)
	}
}
