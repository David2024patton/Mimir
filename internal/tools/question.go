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
// The TUI (Charmbracelet Huh-style selector) or GUI renders the options as interactive
// choices; a "type your own answer" option is always added.
type QuestionTool struct {
	// AskFunc, when set by the TUI or HTTP server, blocks to let the user pick
	// an option interactively and returns their formatted answer.
	AskFunc func(ctx context.Context, questions []Question) (string, error)
}

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
	rawQs, _ := args["questions"].([]any)
	questions := parseQuestions(rawQs)
	if t.AskFunc != nil && len(questions) > 0 {
		return t.AskFunc(ctx, questions)
	}
	return fmt.Sprintf("posed %d question(s) to the user (awaiting selection)", len(rawQs)), nil
}

func parseQuestions(raw []any) []Question {
	var out []Question
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		q := Question{
			Question: strField(m, "question"),
			Header:   strField(m, "header"),
		}
		if mult, ok := m["multiple"].(bool); ok {
			q.Multiple = mult
		}
		if opts, ok := m["options"].([]any); ok {
			for _, ro := range opts {
				if om, ok := ro.(map[string]any); ok {
					q.Options = append(q.Options, Option{
						Label:       strField(om, "label"),
						Description: strField(om, "description"),
					})
				}
			}
		}
		if q.Question != "" {
			out = append(out, q)
		}
	}
	return out
}

func strField(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}
