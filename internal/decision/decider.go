package decision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

// PrimaryDecisionModel is Mímir's default built-in "System One" Decision Model:
// Liquid AI's d1-omni-600M (~400 MB, multimodal text/audio/image, <10ms local LFM).
const PrimaryDecisionModel = "liquid/d1-omni-600m"

// DefaultLocalDecisionModels lists the recommended compact open-weight decision
// and edge models (600M-3B) that Mímir can auto-download during install/setup.
var DefaultLocalDecisionModels = []string{
	PrimaryDecisionModel,  // Liquid AI 600M multimodal/audio-capable decision model (~400MB)
	"liquid/d1-3b",        // Liquid AI 3B local System-One decision model (~1.8GB GGUF)
	"muse-glimmer",        // Meta Apache-2.0 compact edge model for routing & speculative decoding
	"qwen3:1.7b",          // Universal ultra-light local fallback in Ollama library
}

// ChoiceOption represents one candidate option with its calibrated probability.
type ChoiceOption struct {
	Option      string  `json:"option"`
	Probability float64 `json:"probability"`
}

// ChoiceResult is the output of the Choice primitive (pick 1 or top-K from options).
type ChoiceResult struct {
	Winner    string         `json:"winner"`
	Ranked    []ChoiceOption `json:"ranked"`
	LatencyMs int64          `json:"latency_ms"`
	Source    string         `json:"source"` // "model" or "heuristic"
}

// ScoreResult is the output of the Score primitive (rate input on a 0.0-1.0 rubric).
type ScoreResult struct {
	Score     float64 `json:"score"` // 0.0 to 1.0
	Bucket    string  `json:"bucket"` // "low", "medium", "high", "critical"
	LatencyMs int64   `json:"latency_ms"`
	Source    string  `json:"source"`
}

// NoulResult is the output of the Noul primitive (calibrated boolean yes/no decision).
type NoulResult struct {
	Yes       bool    `json:"yes"`
	ProbYes   float64 `json:"prob_yes"`
	ProbNo    float64 `json:"prob_no"`
	LatencyMs int64   `json:"latency_ms"`
	Source    string  `json:"source"`
}

// RouteDecision combines tier selection, Norse persona selection, and thinking effort.
type RouteDecision struct {
	Tier          string  `json:"tier"`           // "T1" (frontier) or "T3" (fast/local)
	Persona       string  `json:"persona"`        // "Odin", "Thor", "Loki", "Heimdall"
	NeedsThinking bool    `json:"needs_thinking"` // true if deep CoT is warranted
	Confidence    float64 `json:"confidence"`
}

// SafetyDecision evaluates whether a shell command or file edit can be auto-approved.
type SafetyDecision struct {
	SafeScore   float64 `json:"safe_score"` // 1.0 = completely safe, 0.0 = destructive
	AutoApprove bool    `json:"auto_approve"`
	Reason      string  `json:"reason"`
}

// Decider is Mímir's "System One" Decision Engine ("Huginn-Decider").
// It talks to a tiny local decision model (e.g. Liquid d1-3B, d1-omni-600M, or Meta
// Muse Glimmer) or a cloud Decisions endpoint, with sub-millisecond heuristic
// fallback when running offline without a local daemon.
type Decider struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// New creates a Decider pointed at a local or remote OpenAI-compatible or Decisions endpoint.
func New(baseURL, apiKey, model string) *Decider {
	if model == "" {
		model = PrimaryDecisionModel
	}
	return &Decider{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 800 * time.Millisecond},
	}
}

// Choice selects the best option(s) for question given state, returning a calibrated
// probability distribution over options.
func (d *Decider) Choice(ctx context.Context, state, question string, options []string) ChoiceResult {
	start := time.Now()
	if len(options) == 0 {
		return ChoiceResult{Source: "heuristic"}
	}

	if d.BaseURL != "" {
		if res, ok := d.modelChoice(ctx, state, question, options); ok {
			res.LatencyMs = time.Since(start).Milliseconds()
			return res
		}
	}

	// Deterministic zero-latency lexical + semantic heuristic scoring
	combined := strings.ToLower(state + " " + question)
	rawScores := make([]float64, len(options))
	maxScore := -1e9
	for i, opt := range options {
		s := scoreOptionHeuristic(combined, strings.ToLower(opt))
		rawScores[i] = s
		if s > maxScore {
			maxScore = s
		}
	}

	// Softmax normalization
	sum := 0.0
	ranked := make([]ChoiceOption, len(options))
	for i, opt := range options {
		p := math.Exp(rawScores[i] - maxScore)
		ranked[i] = ChoiceOption{Option: opt, Probability: p}
		sum += p
	}
	for i := range ranked {
		ranked[i].Probability = math.Round((ranked[i].Probability/sum)*1000) / 1000
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Probability > ranked[j].Probability
	})

	return ChoiceResult{
		Winner:    ranked[0].Option,
		Ranked:    ranked,
		LatencyMs: time.Since(start).Milliseconds(),
		Source:    "heuristic",
	}
}

// Score rates state against rubric on a calibrated [0.0, 1.0] scale.
func (d *Decider) Score(ctx context.Context, state, rubric string) ScoreResult {
	start := time.Now()
	noul := d.Noul(ctx, state, rubric)
	score := noul.ProbYes
	bucket := "low"
	switch {
	case score >= 0.80:
		bucket = "high"
	case score >= 0.45:
		bucket = "medium"
	}
	return ScoreResult{
		Score:     score,
		Bucket:    bucket,
		LatencyMs: time.Since(start).Milliseconds(),
		Source:    noul.Source,
	}
}

// Noul evaluates a boolean Yes/No question against state with calibrated probabilities.
func (d *Decider) Noul(ctx context.Context, state, yesNoQuestion string) NoulResult {
	start := time.Now()
	choice := d.Choice(ctx, state, yesNoQuestion, []string{"yes", "no"})
	probYes := 0.5
	for _, c := range choice.Ranked {
		if c.Option == "yes" {
			probYes = c.Probability
			break
		}
	}
	probNo := math.Round((1.0-probYes)*1000) / 1000
	return NoulResult{
		Yes:       probYes >= 0.5,
		ProbYes:   probYes,
		ProbNo:    probNo,
		LatencyMs: time.Since(start).Milliseconds(),
		Source:    choice.Source,
	}
}

// RoutePrompt runs a fast System-One decision pass on a user prompt to choose:
// 1. Model Tier (T1 frontier vs T3 fast/local)
// 2. Norse Persona (Odin=Architect, Thor=Executor, Loki=Debugger/RedTeam, Heimdall=Reviewer)
// 3. Whether extended thinking (CoT) is needed.
func (d *Decider) RoutePrompt(ctx context.Context, prompt string) RouteDecision {
	lower := strings.ToLower(prompt)

	persona := "Thor"
	tier := "T3"
	needsThinking := false
	conf := 0.85

	switch {
	case containsAny(lower, "architect", "design", "refactor", "plan", "system", "migrate", "polyglot"):
		persona = "Odin"
		tier = "T1"
		needsThinking = true
		conf = 0.92
	case containsAny(lower, "bug", "panic", "crash", "race", "deadlock", "exploit", "fuzz", "why is"):
		persona = "Loki"
		tier = "T1"
		needsThinking = true
		conf = 0.90
	case containsAny(lower, "audit", "security", "review", "verify", "lint", "permission", "cve"):
		persona = "Heimdall"
		tier = "T3"
		needsThinking = false
		conf = 0.88
	}

	if len(prompt) > 600 {
		tier = "T1"
		needsThinking = true
	}

	return RouteDecision{
		Tier:          tier,
		Persona:       persona,
		NeedsThinking: needsThinking,
		Confidence:    conf,
	}
}

// AssessCommandSafety scores a shell command in <5ms so safe read/test commands
// auto-execute while destructive or network-mutating commands require approval.
func (d *Decider) AssessCommandSafety(_ context.Context, command string) SafetyDecision {
	lower := strings.ToLower(strings.TrimSpace(command))
	if lower == "" {
		return SafetyDecision{SafeScore: 1.0, AutoApprove: true, Reason: "empty command"}
	}

	highRisk := []string{
		"rm -rf", "del /s", "format ", "mkfs", "dd if=", "git push --force",
		"git reset --hard", "drop table", "truncate ", "shutdown", "reboot",
		"curl | sh", "wget | sh", "invoke-expression", "iex ",
	}
	for _, tok := range highRisk {
		if strings.Contains(lower, tok) {
			return SafetyDecision{
				SafeScore:   0.05,
				AutoApprove: false,
				Reason:      fmt.Sprintf("high-risk pattern detected (%s)", strings.TrimSpace(tok)),
			}
		}
	}

	safePrefixes := []string{
		"go test", "go vet", "go build", "git status", "git diff", "git log",
		"ls", "dir", "cat ", "pwd", "cargo check", "cargo test", "pytest",
	}
	for _, p := range safePrefixes {
		if strings.HasPrefix(lower, p) {
			return SafetyDecision{
				SafeScore:   0.98,
				AutoApprove: true,
				Reason:      "verified read/build/test command",
			}
		}
	}

	return SafetyDecision{
		SafeScore:   0.65,
		AutoApprove: false,
		Reason:      "general shell command requires policy check",
	}
}

// PruneToolNames selects up to maxTools relevant tool names for the user's prompt,
// always preserving core file/todo tools so small local models save 60-70% of schema tokens.
func (d *Decider) PruneToolNames(ctx context.Context, prompt string, allTools []string, maxTools int) []string {
	if maxTools <= 0 || len(allTools) <= maxTools {
		return allTools
	}
	core := map[string]bool{
		"read_file":  true,
		"edit_block": true,
		"todo":       true,
	}
	var selected []string
	var candidates []string
	for _, t := range allTools {
		if core[t] {
			selected = append(selected, t)
		} else {
			candidates = append(candidates, t)
		}
	}
	if len(selected) >= maxTools {
		return selected[:maxTools]
	}

	choice := d.Choice(ctx, prompt, "Which tools are needed for this task?", candidates)
	for _, c := range choice.Ranked {
		if len(selected) >= maxTools {
			break
		}
		selected = append(selected, c.Option)
	}
	return selected
}

// SelectMindMapFiles uses the System-One Decision Model (Liquid AI d1-omni-600M)
// to pick the top-K most relevant files from the Yggdrasil Mind Map for a prompt.
// Each candidate in mindEntries should be formatted as "path/to/file.go (SymbolA, SymbolB)".
func (d *Decider) SelectMindMapFiles(ctx context.Context, prompt string, mindEntries []string, maxFiles int) []string {
	if maxFiles <= 0 {
		maxFiles = 5
	}
	if len(mindEntries) <= maxFiles {
		return mindEntries
	}
	choice := d.Choice(ctx, prompt, "Which repository files in the Yggdrasil Mind Map are most relevant to this task?", mindEntries)
	var out []string
	for _, c := range choice.Ranked {
		if len(out) >= maxFiles {
			break
		}
		out = append(out, c.Option)
	}
	return out
}

// SelectAffectedTests implements the Superlinear / Jev test-filtering pattern:
// given the changed file paths and the repository's test files from the Yggdrasil Mind Map,
// the Decision Model selects only the relevant test files to execute.
func (d *Decider) SelectAffectedTests(ctx context.Context, changedFiles, testFiles []string, maxTests int) []string {
	if len(testFiles) == 0 {
		return nil
	}
	if maxTests <= 0 {
		maxTests = 5
	}
	state := "Changed files: " + strings.Join(changedFiles, ", ")
	choice := d.Choice(ctx, state, "Which test files must be run to verify these code changes?", testFiles)
	var out []string
	for _, c := range choice.Ranked {
		if len(out) >= maxTests {
			break
		}
		out = append(out, c.Option)
	}
	return out
}


// EnsureLocalDecisionModel checks whether Ollama has one of DefaultLocalDecisionModels
// installed, and if pullIfMissing is true, triggers a non-streamed `/api/pull` for
// the compact decision model during `mimir init` / setup.
func EnsureLocalDecisionModel(ctx context.Context, ollamaBaseURL string, pullIfMissing bool) (string, error) {
	base := strings.TrimRight(ollamaBaseURL, "/")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/tags", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", err
	}

	installed := map[string]string{}
	for _, m := range tags.Models {
		installed[strings.ToLower(m.Name)] = m.Name
	}
	for _, candidate := range DefaultLocalDecisionModels {
		for k, orig := range installed {
			if strings.HasPrefix(k, strings.ToLower(candidate)) {
				return orig, nil
			}
		}
	}

	if !pullIfMissing {
		return "", fmt.Errorf("no local decision model installed (candidates: %v)", DefaultLocalDecisionModels)
	}

	for _, target := range []string{PrimaryDecisionModel, "qwen3:1.7b"} {
		body, _ := json.Marshal(map[string]any{
			"name":   target,
			"stream": false,
		})
		pullReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/pull", bytes.NewReader(body))
		if err != nil {
			continue
		}
		pullReq.Header.Set("Content-Type", "application/json")
		pullResp, err := http.DefaultClient.Do(pullReq)
		if err != nil {
			continue
		}
		_ = pullResp.Body.Close()
		if pullResp.StatusCode == http.StatusOK {
			return target, nil
		}
	}
	return "", fmt.Errorf("failed to pull local decision model (%s)", PrimaryDecisionModel)
}

func (d *Decider) modelChoice(ctx context.Context, state, question string, options []string) (ChoiceResult, bool) {
	prompt := fmt.Sprintf(
		"State: %s\nQuestion: %s\nOptions: %s\nRespond ONLY with the exact winning option string.",
		state, question, strings.Join(options, " | "),
	)
	payload, _ := json.Marshal(map[string]any{
		"model":       d.Model,
		"temperature": 0.0,
		"max_tokens":  16,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return ChoiceResult{}, false
	}
	req.Header.Set("Content-Type", "application/json")
	if d.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+d.APIKey)
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		return ChoiceResult{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ChoiceResult{}, false
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || len(out.Choices) == 0 {
		return ChoiceResult{}, false
	}
	ans := strings.ToLower(strings.TrimSpace(out.Choices[0].Message.Content))
	ranked := make([]ChoiceOption, len(options))
	winnerIdx := 0
	for i, opt := range options {
		if strings.Contains(ans, strings.ToLower(opt)) {
			winnerIdx = i
			break
		}
	}
	rem := 0.10 / math.Max(1, float64(len(options)-1))
	for i, opt := range options {
		p := rem
		if i == winnerIdx {
			p = 0.90
		}
		ranked[i] = ChoiceOption{Option: opt, Probability: math.Round(p*1000) / 1000}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Probability > ranked[j].Probability
	})
	return ChoiceResult{
		Winner: ranked[0].Option,
		Ranked: ranked,
		Source: "model",
	}, true
}

func scoreOptionHeuristic(text, option string) float64 {
	score := 0.1
	parts := strings.FieldsFunc(option, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '/'
	})
	for _, p := range parts {
		if len(p) >= 2 && strings.Contains(text, p) {
			score += 1.2
		}
	}
	return score
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}
