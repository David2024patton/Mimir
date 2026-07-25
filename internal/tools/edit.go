package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EditBlockTool does a targeted search/replace in a file (F3.4). It tries an exact
// match first, then a whitespace-normalized fallback. expected_replacements controls
// how many occurrences to replace (default 1; 0 = all).
type EditBlockTool struct{}

func (t *EditBlockTool) Name() string { return "edit_block" }
func (t *EditBlockTool) Description() string {
	return "Replace an exact block of text in a file. Use old_string/new_string; set expected_replacements to 0 to replace all occurrences."
}
func (t *EditBlockTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":                 map[string]any{"type": "string"},
			"old_string":           map[string]any{"type": "string"},
			"new_string":           map[string]any{"type": "string"},
			"expected_replacements": map[string]any{"type": "integer", "description": "0 = replace all; default 1"},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

func (t *EditBlockTool) Run(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	oldStr, _ := args["old_string"].(string)
	newStr, _ := args["new_string"].(string)
	if oldStr == newStr {
		return "", fmt.Errorf("old_string and new_string are identical")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)

	n := 1
	if v, ok := args["expected_replacements"].(float64); ok {
		n = int(v)
	}

	count := strings.Count(content, oldStr)
	if count == 0 {
		// Whitespace-normalized fallback: collapse runs of spaces/tabs.
		norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
		nOld, nNew, nContent := norm(oldStr), norm(newStr), norm(content)
		if c := strings.Count(nContent, nOld); c > 0 {
			content = replaceN(nContent, nOld, nNew, n)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return "", err
			}
			return fmt.Sprintf("edited %s (normalized fallback, %d replacement(s))", path, min(n, c)), nil
		}
		return "", fmt.Errorf("old_string not found in %s", path)
	}
	if n == 0 {
		content = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		if count < n {
			return "", fmt.Errorf("expected %d occurrence(s) of old_string but found %d in %s", n, count, path)
		}
		content = replaceN(content, oldStr, newStr, n)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	replaced := n
	if n == 0 {
		replaced = count
	}
	return fmt.Sprintf("edited %s (%d replacement(s))", path, replaced), nil
}

func replaceN(s, old, new string, n int) string {
	var b strings.Builder
	i := 0
	for k := 0; k < n; k++ {
		j := strings.Index(s[i:], old)
		if j < 0 {
			break
		}
		b.WriteString(s[i : i+j])
		b.WriteString(new)
		i += j + len(old)
	}
	b.WriteString(s[i:])
	return b.String()
}
