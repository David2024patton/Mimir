package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/David2024patton/Mimir/internal/llm"
)

// HandoffState captures the structured context needed for cross-session continuity.
type HandoffState struct {
	SessionID        string
	Timestamp        time.Time
	CurrentObjective string
	CompletedTasks   []string
	VerifiedFiles    []string
	RemainingTodos   []string
	FailedApproaches []string
	NextCommand      string
}

// GenerateHandoffMarkdown analyzes conversation history, todo items, and Yggdrasil
// mind map status to produce a structured .mimir/HANDOFF.md document.
func GenerateHandoffMarkdown(sessionID string, history []llm.Message, todoLines []string, mindMapSummary string) string {
	state := HandoffState{
		SessionID: sessionID,
		Timestamp: time.Now().UTC(),
	}

	// 1. Extract Current Objective from earliest user message
	for _, m := range history {
		if m.Role == "user" && !strings.HasPrefix(m.Content, "[") {
			state.CurrentObjective = strings.TrimSpace(m.Content)
			break
		}
	}
	if state.CurrentObjective == "" {
		state.CurrentObjective = "Ongoing software development sprint."
	}

	// 2. Parse Todo Checklist into completed vs remaining
	for _, line := range todoLines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "[completed]") || strings.Contains(trimmed, "[x]") {
			state.CompletedTasks = append(state.CompletedTasks, trimmed)
		} else {
			state.RemainingTodos = append(state.RemainingTodos, trimmed)
		}
	}

	// 3. Scan tool history for touched files and failed approaches
	seenFiles := make(map[string]bool)
	for _, m := range history {
		if m.Role == "tool" {
			if strings.Contains(strings.ToLower(m.Content), "error") || strings.Contains(strings.ToLower(m.Content), "failed") {
				errMsg := strings.TrimSpace(m.Content)
				if len(errMsg) > 160 {
					errMsg = errMsg[:160] + "..."
				}
				state.FailedApproaches = append(state.FailedApproaches, errMsg)
			}
		}
		for _, tc := range m.ToolCalls {
			if tc.Name == "write_file" || tc.Name == "edit_block" {
				target := extractTargetFile(tc.Arguments)
				if target != "" && !seenFiles[target] {
					seenFiles[target] = true
					state.VerifiedFiles = append(state.VerifiedFiles, target)
				}
			}
		}
	}

	// 4. Determine Suggested Next Command
	if len(state.RemainingTodos) > 0 {
		state.NextCommand = "mimir \"continue working on remaining checklist items\""
	} else {
		state.NextCommand = "mimir review"
	}

	var b strings.Builder
	b.WriteString("# Mímir Session Handoff & Continuity State\n")
	b.WriteString(fmt.Sprintf("- **Session ID**: `%s`\n", state.SessionID))
	b.WriteString(fmt.Sprintf("- **Generated At**: `%s`\n", state.Timestamp.Format(time.RFC3339)))
	if mindMapSummary != "" {
		b.WriteString(fmt.Sprintf("- **Repository State**: %s\n", mindMapSummary))
	}
	b.WriteString("\n")

	b.WriteString("## 1. Current Objective & Architectural State\n")
	b.WriteString(state.CurrentObjective + "\n\n")

	b.WriteString("## 2. Completed Tasks & Verified Files\n")
	if len(state.CompletedTasks) > 0 {
		b.WriteString("### Completed Checklist Items\n")
		for _, t := range state.CompletedTasks {
			b.WriteString(t + "\n")
		}
	}
	if len(state.VerifiedFiles) > 0 {
		b.WriteString("\n### Modified / Verified Files\n")
		for _, f := range state.VerifiedFiles {
			b.WriteString("- `" + f + "`\n")
		}
	}
	if len(state.CompletedTasks) == 0 && len(state.VerifiedFiles) == 0 {
		b.WriteString("No completed tasks or modified files recorded yet.\n")
	}
	b.WriteString("\n")

	b.WriteString("## 3. Remaining Todo Checklist (In-Progress & Pending)\n")
	if len(state.RemainingTodos) > 0 {
		for _, t := range state.RemainingTodos {
			b.WriteString(t + "\n")
		}
	} else {
		b.WriteString("- No pending checklist items. All scheduled tasks completed.\n")
	}
	b.WriteString("\n")

	b.WriteString("## 4. Failed Approaches & Guardrails (Do Not Repeat)\n")
	if len(state.FailedApproaches) > 0 {
		for _, f := range state.FailedApproaches {
			b.WriteString("- " + f + "\n")
		}
	} else {
		b.WriteString("- Zero failed tool executions recorded in this session.\n")
	}
	b.WriteString("\n")

	b.WriteString("## 5. Exact Next Command to Run\n")
	b.WriteString("```bash\n" + state.NextCommand + "\n```\n")

	return b.String()
}

// WriteHandoffFile saves the generated handoff markdown to .mimir/HANDOFF.md.
func WriteHandoffFile(workDir, content string) (string, error) {
	if workDir == "" {
		workDir = "."
	}
	dir := filepath.Join(workDir, ".mimir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	outPath := filepath.Join(dir, "HANDOFF.md")
	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return outPath, nil
}
