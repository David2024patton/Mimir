package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/David2024patton/Mimir/internal/critic"
)

// EpistemicAuditTool exposes Mímir's Epistemic Critic & Verification Node as a
// callable harness tool to audit draft reports/copy against retrieved context.
type EpistemicAuditTool struct{}

func (t *EpistemicAuditTool) Name() string { return "epistemic_audit" }

func (t *EpistemicAuditTool) Description() string {
	return "Run the Epistemic Critic & Verification Node against <retrieved_context> and <draft_report> to flag UNVERIFIED_TRAINING_DATA, EXTRAPOLATION, vendor bias, and hype, and compute pass_audit."
}

func (t *EpistemicAuditTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"retrieved_context": map[string]any{
				"type":        "string",
				"description": "The raw sources, snippets, and tool outputs retrieved during the session",
			},
			"draft_report": map[string]any{
				"type":        "string",
				"description": "The synthesized draft report or copy to audit against retrieved_context",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "Output format: json (default) or markdown",
			},
		},
		"required": []string{"retrieved_context", "draft_report"},
	}
}

func (t *EpistemicAuditTool) Run(_ context.Context, args map[string]any) (string, error) {
	retrieved, _ := args["retrieved_context"].(string)
	draft, _ := args["draft_report"].(string)
	format, _ := args["format"].(string)

	if strings.TrimSpace(draft) == "" {
		return "", fmt.Errorf("draft_report is required")
	}

	report := critic.AuditDeterministic(retrieved, draft)
	if strings.EqualFold(strings.TrimSpace(format), "markdown") {
		return report.FormatMarkdown(), nil
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
