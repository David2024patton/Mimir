package agent

import (
	"strings"
	"testing"
)

func TestPersonaByName(t *testing.T) {
	p, ok := PersonaByName("odin")
	if !ok {
		t.Fatal("odin persona not found")
	}
	if p.Role == "" || p.Prompt == "" {
		t.Errorf("odin persona incomplete: %+v", p)
	}
	if _, ok := PersonaByName("nonexistent"); ok {
		t.Error("expected nonexistent persona to not be found")
	}
}

func TestBuiltinPersonas(t *testing.T) {
	if len(BuiltinPersonas) < 9 {
		t.Errorf("expected at least 9 builtin personas, got %d", len(BuiltinPersonas))
	}
	// every persona must have a name, role, and prompt
	for _, p := range BuiltinPersonas {
		if p.Name == "" || p.Role == "" || p.Prompt == "" {
			t.Errorf("persona missing fields: %+v", p)
		}
	}
}

func TestDefaultVoiceAndFilter(t *testing.T) {
	p, ok := PersonaByName("odin")
	if !ok {
		t.Fatal("odin persona not found")
	}
	full := p.FullPrompt()
	if !strings.Contains(full, "10th-grade") {
		t.Error("FullPrompt should include the 10th-grade voice")
	}
	if !strings.Contains(full, "NEVER use em dashes") {
		t.Error("FullPrompt should include the no-em-dash rule")
	}
	if !strings.Contains(full, "Always use the to-do list") {
		t.Error("FullPrompt should include the always-use-todo rule")
	}
	if !strings.Contains(full, "question tool") {
		t.Error("FullPrompt should mention the question tool")
	}
	out := StripEmDashes("hello \u2014 world \u2013 ok")
	if out != "hello - world - ok" {
		t.Errorf("StripEmDashes = %q, want %q", out, "hello - world - ok")
	}
}
