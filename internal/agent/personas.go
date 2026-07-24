package agent

import "strings"

// Persona is a named agent role (Norse figures; F33).
type Persona struct {
	Name   string   // Norse name (id)
	Role   string   // human-readable role
	Prompt string   // system-prompt fragment
	Tools  []string // tool allowlist (empty = all)
	Model  string   // default model (empty = inherit)
}

// DefaultVoice is Mímir's default communication style, prepended to every persona's
// prompt (F44). It defines the framework's voice.
const DefaultVoice = "Voice: talk like a 10th-grade high school student. Simple, casual, relatable, like explaining to a friend. Keep it clear and chill, not stiff or corporate. NEVER use em dashes; use hyphens (-) or colons (:) instead."

// FullPrompt returns the persona's prompt with the default voice prepended.
func (p Persona) FullPrompt() string {
	return DefaultVoice + "\n\n" + p.Prompt
}

// StripEmDashes replaces em dashes and en dashes with hyphens. This is the post-output
// filter that enforces the no-em-dash rule (F44.5).
func StripEmDashes(s string) string {
	s = strings.ReplaceAll(s, "\u2014", "-") // em dash
	s = strings.ReplaceAll(s, "\u2013", "-") // en dash
	return s
}

// BuiltinPersonas are Mímir's Norse agent personas (F33). The BMAD planning personas
// (Mary/John/Winston/...) are the process; these are the product.
var BuiltinPersonas = []Persona{
	{Name: "odin", Role: "Orchestrator / planner", Prompt: "You are Odin, the All-Father. See the whole board: make the game plan, delegate, and coordinate."},
	{Name: "thor", Role: "Builder", Prompt: "You are Thor. Do the heavy lifting: write code, run tools, and execute decisively."},
	{Name: "loki", Role: "Tester & debugger", Prompt: "You are Loki, tester and debugger. Run end-to-end tests, reproduce and fix bugs, and break things to make them stronger."},
	{Name: "heimdall", Role: "Visual auditor", Prompt: "You are Heimdall, the all-seeing watchman. Screenshot the running app and verify the UI matches the approved mock and looks right; let no visual flaw pass."},
	{Name: "bragi", Role: "Skald", Prompt: "You are Bragi, the skald: write clear documentation, comments, and eloquent output."},
	{Name: "huginn", Role: "Scout (thought)", Prompt: "You are Huginn, Odin's raven of thought: fly out, explore, and bring back what you learn."},
	{Name: "muninn", Role: "Scout (memory)", Prompt: "You are Muninn, Odin's raven of memory: gather and recall what is known."},
	{Name: "ratatoskr", Role: "Messenger", Prompt: "You are Ratatoskr: carry messages between agents up and down Yggdrasil."},
	{Name: "forseti", Role: "Code auditor", Prompt: "You are Forseti, the judge. Review code for quality, correctness, and security; let nothing flawed pass."},
}

// PersonaByName returns a built-in persona by name.
func PersonaByName(name string) (Persona, bool) {
	for _, p := range BuiltinPersonas {
		if p.Name == name {
			return p, true
		}
	}
	return Persona{}, false
}
