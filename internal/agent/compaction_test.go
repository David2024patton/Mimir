package agent

import (
	"fmt"
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
