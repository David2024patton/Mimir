package tools

import (
	"context"
	"fmt"
)

// Option is one choice in a question.
type Option struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// Question is a structured question posed to the user (like opencode's question tool).
type Question struct {
	Question string   `json:"question"`
	Header   string   `json:"header"`
	Options  []Option `json:"options"`
	Multiple bool     `json:"multiple"`
}

// QuestionTool lets the agent ask the user structured multiple-choice questions (F45).
// The GUI/chat renders the options as clickable choices; a "type your own answer"
// option is always added. Used in Discovery and for any decision/clarification.
type QuestionTool struct{}

func (t *QuestionTool) Name() string { return "question" }
func (t *QuestionTool) Description() string {
	return "Ask the user a structured question with options to gather preferences, decisions, or clarifications. Use during Discovery and whenever you need the user to choose."
}
func (t *QuestionTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"questions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{"type": "string"},
						"header":   map[string]any{"type": "string"},
						"multiple": map[string]any{"type": "boolean"},
						"options": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"label":       map[string]any{"type": "string"},
									"description": map[string]any{"type": "string"},
								},
								"required": []string{"label", "description"},
							},
						},
					},
					"required": []string{"question", "header", "options"},
				},
			},
		},
		"required": []string{"questions"},
	}
}

func (t *QuestionTool) Run(ctx context.Context, args map[string]any) (string, error) {
	// The GUI/chat renders the questions and collects the user's selections (F45 + F9).
	// Skeleton placeholder: the real implementation wires to the interactive flow.
	qs, _ := args["questions"].([]any)
	return fmt.Sprintf("posed %d question(s) to the user (awaiting selection)", len(qs)), nil
}
