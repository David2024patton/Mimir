package agent

import (
	"strings"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
)

// SmolPolicy controls whether SmolMode is auto-activated ("auto"), forced on ("on"),
// or forced off ("off"). Default is "auto".
const (
	SmolAuto = "auto"
	SmolOn   = "on"
	SmolOff  = "off"
)

var smolTokens = []string{
	"0.5b", "600m", "1b", "1.5b", "1.7b", "2b", "3b", "4b", "6.7b", "7b", "8b", "9b",
	"12b", "13b", "14b", "16b", "22b", "24b", "27b", "30b", "32b",
	"mini", "flash", "smol", "nano", "haiku", "small", "phi-", "gemma",
	"muse-glimmer", "d1-", "clef-flash",
}

var largeLocalTokens = []string{
	"70b", "72b", "110b", "120b", "123b", "235b", "405b", "671b",
}

// DetectSmolModel returns true when the active model is a small/local model (<=32B)
// or a fast/compact cloud tier (flash/mini/nano/haiku), triggering automatic
// smolcoder optimizations (lean 2-paragraph system prompt, aggressive 2-turn
// history placeholder pruning, and automatic /no_think on mechanical tool steps).
func DetectSmolModel(providerID, model string, local bool) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	p := strings.ToLower(strings.TrimSpace(providerID))

	for _, tok := range smolTokens {
		if strings.Contains(m, tok) {
			return true
		}
	}
	if local || p == "ollama" || p == "lmstudio" || p == "llamacpp" {
		for _, big := range largeLocalTokens {
			if strings.Contains(m, big) {
				return false
			}
		}
		return true
	}
	return false
}

// smolSystemPrompt is the ultra-compact 2-paragraph system prompt (inspired by
// smolcoder.dev) designed specifically for 7B-32B local models and small context windows.
func smolSystemPrompt(memories []cortex.Neuron, schemas []llm.ToolSchema) string {
	var b strings.Builder
	b.WriteString("You are Mímir (⚡SmolMode), a fast, precise coding agent with persistent memory. ")
	b.WriteString("Act directly using your tools: read files before editing, keep edit_block old_string snippets small (under 15 lines), and track multi-step tasks with the todo tool. Never use em dashes.\n\n")

	if len(schemas) > 0 {
		names := make([]string, len(schemas))
		for i, s := range schemas {
			names[i] = s.Name
		}
		b.WriteString("Available tools: " + strings.Join(names, ", ") + ". ")
		b.WriteString("Example tool usage: `read_file` with `{\"path\":\"main.go\"}`, `edit_block` with `{\"path\":\"main.go\",\"old_string\":\"old\",\"new_string\":\"new\"}`, `todo` with `{\"action\":\"add\",\"content\":\"step 1\"}`.")
	}

	if len(memories) > 0 {
		b.WriteString("\n\nRelevant memory:")
		for _, m := range memories {
			b.WriteString("\n- " + m.Content)
		}
	}

	b.WriteString("\n\nWhen asking multiple-choice questions, use:\n[choice]\nA) option\nB) option\n[/choice]")
	return b.String()
}
