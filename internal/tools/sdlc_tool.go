package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/David2024patton/Mimir/internal/sdlc"
)

// SDLCTool allows Mímir to inspect, advance, or save deliverables for the 7-Phase
// Software Development Life Cycle (Planning, Analysis, Design, Implementation,
// Testing, Deployment, Maintenance) across Agile, Waterfall, and DevOps models.
type SDLCTool struct {
	Manager *sdlc.Manager
}

func (t *SDLCTool) Name() string { return "sdlc" }

func (t *SDLCTool) Description() string {
	return "Manage the 7-Phase SDLC (status, set_phase, advance, set_method, save_deliverable) so only the active lifecycle phase prompt & deliverable are loaded."
}

func (t *SDLCTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "One of: status, set_phase, advance, set_method, save_deliverable, sign_off",
			},
			"phase": map[string]any{
				"type":        "string",
				"description": "Phase step (1..7) or ID (planning, analysis, design, implementation, testing, deployment, maintenance)",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "SDLC execution methodology: agile, waterfall, or devops",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Markdown content when action is save_deliverable (writes .mimir/sdlc/0X-*.md and signs off the phase)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *SDLCTool) Run(_ context.Context, args map[string]any) (string, error) {
	if t.Manager == nil {
		t.Manager = sdlc.NewManager(".")
	}
	action, _ := args["action"].(string)
	phaseStr, _ := args["phase"].(string)
	methodStr, _ := args["method"].(string)
	content, _ := args["content"].(string)

	switch strings.ToLower(strings.TrimSpace(action)) {
	case "status", "board", "":
		return t.Manager.FormatBoard(), nil
	case "set_phase", "phase":
		spec, err := t.Manager.SetPhase(phaseStr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("SDLC active phase -> [%d/7] %s (Persona: %s | Deliverable: %s)",
			spec.Step, spec.Title, spec.Persona, spec.Deliverable), nil
	case "advance", "next":
		spec, err := t.Manager.Advance()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Advanced SDLC -> [%d/7] %s (Persona: %s | Deliverable: %s)",
			spec.Step, spec.Title, spec.Persona, spec.Deliverable), nil
	case "set_method", "method":
		if t.Manager.Methodology() == sdlc.MethodWaterfall {
			return "", fmt.Errorf("security policy: SDLC methodology cannot be changed by the model while in strict Waterfall mode (use the CLI or TUI)")
		}
		if err := t.Manager.SetMethodology(methodStr); err != nil {
			return "", err
		}
		return fmt.Sprintf("SDLC methodology set to %q", t.Manager.Methodology()), nil
	case "save_deliverable", "deliverable":
		if strings.TrimSpace(content) == "" {
			return "", fmt.Errorf("content is required for save_deliverable")
		}
		var pid sdlc.PhaseID
		if phaseStr != "" {
			if spec, ok := sdlc.LookupPhase(phaseStr); ok {
				pid = spec.ID
			}
		}
		path, err := t.Manager.SaveDeliverable(pid, content)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Saved SDLC deliverable to %s.", path), nil
	case "sign_off", "signoff":
		if t.Manager.Methodology() == sdlc.MethodWaterfall {
			return "", fmt.Errorf("security policy: sign_off in strict Waterfall mode requires user approval via `mimir sdlc signoff` or TUI /sdlc signoff")
		}
		var pid sdlc.PhaseID
		if phaseStr != "" {
			if spec, ok := sdlc.LookupPhase(phaseStr); ok {
				pid = spec.ID
			}
		}
		spec := t.Manager.SignOff(pid)
		return fmt.Sprintf("Signed off SDLC Step %d (%s - %s).", spec.Step, spec.Title, spec.Deliverable), nil
	default:
		return "", fmt.Errorf("unknown sdlc action %q (use status, set_phase, advance, set_method, save_deliverable, sign_off)", action)
	}
}
