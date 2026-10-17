package agent

import (
	"strings"
	"testing"
)

func TestDetectSmolModel(t *testing.T) {
	cases := []struct {
		provider string
		model    string
		local    bool
		want     bool
	}{
		{"ollama", "qwen3:8b", true, true},
		{"lmstudio", "deepseek-r1-distill-qwen-14b", true, true},
		{"ollama", "llama3.3:70b", true, false},
		{"mimo", "mimo-v2-flash", false, true},
		{"openai", "gpt-4o-mini", false, true},
		{"mimo", "mimo-v2-pro", false, false},
		{"anthropic", "claude-3-7-sonnet", false, false},
	}

	for _, tc := range cases {
		got := DetectSmolModel(tc.provider, tc.model, tc.local)
		if got != tc.want {
			t.Errorf("DetectSmolModel(%q, %q, local=%v) = %v, want %v",
				tc.provider, tc.model, tc.local, got, tc.want)
		}
	}
}

func TestAgentAutoActivatesSmolMode(t *testing.T) {
	ag := New(Config{
		Model: "qwen3:8b",
		Local: true,
	})
	if !ag.SmolMode() {
		t.Fatal("expected SmolMode() to auto-activate for local qwen3:8b")
	}
	ag.SetSmolPolicy(SmolOff)
	if ag.SmolMode() {
		t.Fatal("expected SmolMode() to be false when policy is SmolOff")
	}
	ag.SetSmolPolicy(SmolOn)
	if !ag.SmolMode() {
		t.Fatal("expected SmolMode() to be true when policy is SmolOn")
	}
	prompt := smolSystemPrompt(nil, nil)
	if !strings.Contains(prompt, "SmolMode") {
		t.Errorf("expected smolSystemPrompt to mention SmolMode, got: %s", prompt)
	}
}
