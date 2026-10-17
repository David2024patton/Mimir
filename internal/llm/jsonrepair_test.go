package llm

import (
	"encoding/json"
	"testing"
)

func TestRepairJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string becomes empty object",
			in:   "   ",
			want: "path",
		},
		{
			name: "strips markdown code fence",
			in:   "```json\n{\"path\": \"main.go\"}\n```",
			want: "main.go",
		},
		{
			name: "removes trailing comma in object",
			in:   `{"path": "foo.go", "content": "bar", }`,
			want: "foo.go",
		},
		{
			name: "closes truncated string and object",
			in:   `{"path": "internal/agent/agent.go`,
			want: "internal/agent/agent.go",
		},
		{
			name: "closes nested array and object",
			in:   `{"questions": [{"question": "Which mode?", "header": "Mode", "options": [{"label": "Fast", "description": "Low latency"}`,
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RepairJSON(tc.in)
			if !json.Valid([]byte(got)) {
				t.Fatalf("RepairJSON(%q) produced invalid JSON: %q", tc.in, got)
			}
			var m map[string]any
			if err := json.Unmarshal([]byte(got), &m); err != nil {
				t.Fatalf("Unmarshal(%q) failed: %v", got, err)
			}
			if tc.name != "empty string becomes empty object" && tc.want != "" {
				if p, _ := m["path"].(string); p != tc.want {
					t.Fatalf("got path=%q, want %q", p, tc.want)
				}
			}
		})
	}
}
