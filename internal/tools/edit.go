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
type EditBlockTool struct {
	WorkDir string
}

func (t *EditBlockTool) Name() string { return "edit_block" }
func (t *EditBlockTool) Description() string {
	return "Replace an exact block of text in a file within the workspace. Use old_string/new_string; set expected_replacements to 0 to replace all occurrences."
}
func (t *EditBlockTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":                  map[string]any{"type": "string"},
			"old_string":            map[string]any{"type": "string"},
			"new_string":            map[string]any{"type": "string"},
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
	resolved, err := resolveWithinWorkDir(t.WorkDir, path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(resolved)
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
		// 1. CRLF / LF line-ending normalization fallback (critical on Windows).
		if strings.Contains(content, "\r\n") && !strings.Contains(oldStr, "\r\n") {
			crlfOld := strings.ReplaceAll(oldStr, "\n", "\r\n")
			crlfNew := strings.ReplaceAll(newStr, "\n", "\r\n")
			if c := strings.Count(content, crlfOld); c > 0 {
				if n == 0 {
					content = strings.ReplaceAll(content, crlfOld, crlfNew)
				} else {
					content = replaceN(content, crlfOld, crlfNew, n)
				}
				if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
					return "", err
				}
				return fmt.Sprintf("edited %s (CRLF-normalized, %d replacement(s))", path, min(n, c)), nil
			}
		}

		return "", notFoundErr(path, content, oldStr)
	}
	if n == 0 {
		content = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		if count < n {
			return "", fmt.Errorf("expected %d occurrence(s) of old_string but found %d in %s", n, count, path)
		}
		content = replaceN(content, oldStr, newStr, n)
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return "", err
	}
	replaced := n
	if n == 0 {
		replaced = count
	}
	return fmt.Sprintf("edited %s (%d replacement(s))", path, replaced), nil
}

// notFoundErr builds a self-healing error with smolcoder-style bounded line/offset
// location hints so small and fast models immediately see where the target block
// actually lives without re-reading the entire file.
func notFoundErr(path, content, oldStr string) error {
	hint := closestBlockHint(content, oldStr)
	preview := content
	if r := []rune(preview); len(r) > 1500 {
		preview = string(r[:1500]) + "\n...[truncated]"
	}
	if hint != "" {
		return fmt.Errorf(
			"old_string not found in %s. %s\nDo NOT give up or reply yet: copy the EXACT text from the hint or call read_file on %s, then retry edit_block. Current file preview:\n-----\n%s\n-----",
			path, hint, path, preview)
	}
	return fmt.Errorf(
		"old_string not found in %s. Do NOT give up or reply yet: call read_file on %s and copy the EXACT text (including any accented or special characters, byte-for-byte) into old_string, then retry edit_block. Current file content:\n-----\n%s\n-----",
		path, path, preview)
}

// closestBlockHint finds the line number and character offset in content that best
// matches the first non-empty line of oldStr (inspired by smolcoder's editFile offset hints).
func closestBlockHint(content, oldStr string) string {
	var needle string
	for _, l := range strings.Split(oldStr, "\n") {
		trimmed := strings.TrimSpace(l)
		if len(trimmed) >= 4 {
			needle = trimmed
			break
		}
	}
	if needle == "" {
		return ""
	}
	lines := strings.Split(content, "\n")
	offset := 0
	for i, line := range lines {
		if strings.Contains(line, needle) || strings.Contains(strings.ToLower(line), strings.ToLower(needle)) {
			start := max(0, i-2)
			end := min(len(lines), i+6)
			snippet := strings.Join(lines[start:end], "\n")
			return fmt.Sprintf("Closest match is at line %d (byte offset %d):\n```text\n%s\n```", i+1, offset, snippet)
		}
		offset += len(line) + 1
	}
	return ""
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
