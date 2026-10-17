package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillTool implements progressive SKILL.md disclosure (inspired by microsoft/agent-framework-go
// and htekdev/ai-harness). It searches workspace .mimir/skills and the centralized H:\AI\skills
// library so the agent can pull specialized domain instructions on demand without bloating
// the base system prompt.
type SkillTool struct {
	SearchDirs []string
}

func (t *SkillTool) Name() string { return "load_skill" }
func (t *SkillTool) Description() string {
	return "List or load a specialized SKILL.md instruction file by name (action='list' or action='load')."
}
func (t *SkillTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{"type": "string", "enum": []string{"list", "load"}},
			"name":   map[string]any{"type": "string", "description": "Skill folder name when action='load'"},
		},
		"required": []string{"action"},
	}
}

func (t *SkillTool) dirs() []string {
	if len(t.SearchDirs) > 0 {
		return t.SearchDirs
	}
	return []string{".mimir/skills", `H:\AI\skills`}
}

func (t *SkillTool) Run(ctx context.Context, args map[string]any) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "list":
		var names []string
		seen := map[string]bool{}
		for _, d := range t.dirs() {
			entries, err := os.ReadDir(d)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() || seen[e.Name()] {
					continue
				}
				skillPath := filepath.Join(d, e.Name(), "SKILL.md")
				if _, err := os.Stat(skillPath); err == nil {
					seen[e.Name()] = true
					names = append(names, e.Name())
				}
			}
		}
		if len(names) == 0 {
			return "no skills found", nil
		}
		if len(names) > 60 {
			return fmt.Sprintf("%d skills available (first 60): %s", len(names), strings.Join(names[:60], ", ")), nil
		}
		return "available skills: " + strings.Join(names, ", "), nil
	case "load":
		name, _ := args["name"].(string)
		name = filepath.Clean(strings.TrimSpace(name))
		if name == "" || strings.Contains(name, "..") {
			return "", fmt.Errorf("valid skill name required")
		}
		for _, d := range t.dirs() {
			p := filepath.Join(d, name, "SKILL.md")
			if data, err := os.ReadFile(p); err == nil {
				return string(data), nil
			}
		}
		return "", fmt.Errorf("skill %q not found", name)
	default:
		return "", fmt.Errorf("unknown action %q (use 'list' or 'load')", action)
	}
}
