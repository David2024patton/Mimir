package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/David2024patton/Mimir/internal/llm"
)

const (
	// placeholderMinChars is the minimum size of a file write/edit payload before
	// we replace it with a compact placeholder in older turns (smolcoder pattern).
	placeholderMinChars = 300
	// maxOldToolOutputChars caps older tool outputs in conversation history
	// (sausheong/harness Stage 1 in-place tool output trimming).
	maxOldToolOutputChars = 2000
)

// PruneHistoryPlaceholders returns a copy of msgs where older bulky tool-call arguments
// (write_file, edit_block) and oversized tool outputs are replaced with compact
// placeholders (inspired by smolcoder's isHistoryPlaceholder and sausheong/harness Stage 1).
// The most recent keepRecent messages are left untouched so active context stays intact.
func PruneHistoryPlaceholders(msgs []llm.Message, keepRecent int) []llm.Message {
	if keepRecent <= 0 {
		keepRecent = 4
	}
	if len(msgs) <= keepRecent {
		return msgs
	}
	cutoff := len(msgs) - keepRecent
	out := make([]llm.Message, len(msgs))
	copy(out, msgs)

	for i := 0; i < cutoff; i++ {
		m := out[i]
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			tcs := make([]llm.ToolCall, len(m.ToolCalls))
			copy(tcs, m.ToolCalls)
			for j, tc := range tcs {
				if tc.Name == "write_file" || tc.Name == "edit_block" {
					tcs[j].Arguments = compactFileToolArgs(tc.Name, tc.Arguments)
				}
			}
			m.ToolCalls = tcs
			out[i] = m
		} else if m.Role == "tool" && len(m.Content) > maxOldToolOutputChars {
			head := m.Content[:1000]
			tail := m.Content[len(m.Content)-600:]
			omitted := len(m.Content) - 1600
			m.Content = fmt.Sprintf("%s\n...[%d chars compacted from earlier tool output]...\n%s", head, omitted, tail)
			out[i] = m
		}
	}
	return out
}

func compactFileToolArgs(toolName, rawArgs string) string {
	if len(rawArgs) < placeholderMinChars {
		return rawArgs
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(llm.RepairJSON(rawArgs)), &args); err != nil {
		return rawArgs
	}
	path, _ := args["path"].(string)
	if path == "" {
		path = "<file>"
	}
	modified := false
	if content, ok := args["content"].(string); ok && len(content) >= placeholderMinChars {
		args["content"] = fmt.Sprintf("[%d characters already applied to %s. Call read_file for current code.]", len(content), path)
		modified = true
	}
	if newStr, ok := args["new_string"].(string); ok && len(newStr) >= placeholderMinChars {
		args["new_string"] = fmt.Sprintf("[%d characters already applied to %s. Call read_file for current code.]", len(newStr), path)
		modified = true
	}
	if oldStr, ok := args["old_string"].(string); ok && len(oldStr) >= placeholderMinChars {
		args["old_string"] = fmt.Sprintf("[%d characters matched in %s]", len(oldStr), path)
		modified = true
	}
	if !modified {
		return rawArgs
	}
	b, err := json.Marshal(args)
	if err != nil {
		return rawArgs
	}
	return string(b)
}

// EstimateTokens returns a fast token estimate (~4 characters per token) for status
// bars and context budget tracking.
func EstimateTokens(msgs []llm.Message) int {
	chars := 0
	for _, m := range msgs {
		chars += len(m.Content) + 16
		for _, tc := range m.ToolCalls {
			chars += len(tc.Name) + len(tc.Arguments) + 12
		}
	}
	if chars == 0 {
		return 0
	}
	return (chars + 3) / 4
}

// FormatTodoChecklist formats a concise harness-owned checklist block to inject
// into the system prompt so the model never loses track of its plan across compactions.
func FormatTodoChecklist(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return "\n\nActive Todo Checklist (survives context compaction):\n" + strings.Join(items, "\n")
}
