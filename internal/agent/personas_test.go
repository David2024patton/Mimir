package agent

import "testing"

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
