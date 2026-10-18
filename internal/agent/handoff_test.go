package agent

import (
	"os"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/llm"
)

func TestGenerateHandoffMarkdown(t *testing.T) {
	history := []llm.Message{
		{Role: "user", Content: "Build the Level 1 harness lifecycle hooks"},
		{
			Role: "assistant",
			Content: "Editing hooks.go",
			ToolCalls: []llm.ToolCall{
				{
					Name:      "write_file",
					Arguments: `{"path": "internal/agent/hooks.go"}`,
				},
			},
		},
		{Role: "tool", Content: "error: file not found on initial read"},
	}

	todos := []string{
		"- #1 [completed] Implement hooks.go",
		"- #2 [in_progress] Implement handoff.go",
	}

	md := GenerateHandoffMarkdown("session-42", history, todos, "108 files (~237k tokens)")

	if !strings.Contains(md, "session-42") {
		t.Errorf("expected session-42 in handoff, got:\n%s", md)
	}
	if !strings.Contains(md, "Build the Level 1 harness lifecycle hooks") {
		t.Errorf("expected objective in handoff, got:\n%s", md)
	}
	if !strings.Contains(md, "internal/agent/hooks.go") {
		t.Errorf("expected modified file in handoff, got:\n%s", md)
	}
	if !strings.Contains(md, "Implement handoff.go") {
		t.Errorf("expected pending todo in handoff, got:\n%s", md)
	}
	if !strings.Contains(md, "error: file not found on initial read") {
		t.Errorf("expected failed approach recorded, got:\n%s", md)
	}
}

func TestWriteHandoffFile(t *testing.T) {
	dir := t.TempDir()
	content := "# Test Handoff Content"
	path, err := WriteHandoffFile(dir, content)
	if err != nil {
		t.Fatalf("WriteHandoffFile: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil || string(read) != content {
		t.Fatalf("mismatch reading handoff file: %s (%v)", string(read), err)
	}
}
