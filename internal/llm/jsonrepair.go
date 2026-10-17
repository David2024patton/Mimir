package llm

import (
	"encoding/json"
	"strings"
)

// RepairJSON repairs common malformed JSON emitted by smaller local models (7B-32B)
// and fast streaming tool calls (inspired by charmbracelet/fantasy jsonrepair).
// It strips markdown code fences, removes trailing commas before '}' or ']',
// and balances unclosed quotes, brackets, and braces.
func RepairJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "{}"
	}

	// Fast path: already valid JSON.
	if json.Valid([]byte(s)) {
		return s
	}

	// 1. Strip markdown code fences (e.g. ```json ... ```)
	if strings.HasPrefix(s, "```") {
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			s = s[nl+1:]
		} else {
			s = strings.TrimPrefix(s, "```json")
			s = strings.TrimPrefix(s, "```")
		}
		if end := strings.LastIndex(s, "```"); end >= 0 {
			s = s[:end]
		}
		s = strings.TrimSpace(s)
		if json.Valid([]byte(s)) {
			return s
		}
	}

	// 2. Scan character by character to remove trailing commas and track open delimiters.
	var out strings.Builder
	out.Grow(len(s) + 8)

	var stack []byte
	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inString {
			out.WriteByte(ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
			out.WriteByte(ch)
		case '{':
			stack = append(stack, '}')
			out.WriteByte(ch)
		case '[':
			stack = append(stack, ']')
			out.WriteByte(ch)
		case '}', ']':
			if len(stack) > 0 && stack[len(stack)-1] == ch {
				stack = stack[:len(stack)-1]
			}
			out.WriteByte(ch)
		case ',':
			// Skip trailing comma if the next non-whitespace character closes an object/array
			// or if we are at the end of the string.
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
				j++
			}
			if j >= len(s) || s[j] == '}' || s[j] == ']' {
				continue
			}
			out.WriteByte(ch)
		default:
			out.WriteByte(ch)
		}
	}

	// 3. Close dangling escape or string quote.
	if inString {
		if escaped {
			out.WriteByte('\\')
		}
		out.WriteByte('"')
	}

	// 4. Close any unclosed objects or arrays in reverse order.
	for i := len(stack) - 1; i >= 0; i-- {
		out.WriteByte(stack[i])
	}

	repaired := out.String()
	if json.Valid([]byte(repaired)) {
		return repaired
	}
	return s
}
