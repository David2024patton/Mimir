package sdlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PhaseRecord tracks sign-off status and deliverable path for a single SDLC phase.
type PhaseRecord struct {
	Phase       PhaseID   `json:"phase"`
	SignedOff   bool      `json:"signed_off"`
	Deliverable string    `json:"deliverable,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// Manager coordinates the 7-Phase SDLC state machine, methodology rules
// (Agile, Waterfall, DevOps), deliverable artifacts (.mimir/sdlc/), and
// phase-scoped prompt/tool loading so the main prompt remains tiny.
type Manager struct {
	mu         sync.RWMutex
	workDir    string
	current    PhaseID
	method     Methodology
	autoDetect bool
	records    map[PhaseID]PhaseRecord
}

// NewManager creates an SDLC Manager rooted at workDir.
// Defaults to MethodAgile, PhaseImplementation, and autoDetect=true so each
// user request automatically loads only the SDLC phase micro-prompt it needs.
func NewManager(workDir string) *Manager {
	if workDir == "" {
		workDir = "."
	}
	return &Manager{
		workDir:    workDir,
		current:    PhaseImplementation,
		method:     MethodAgile,
		autoDetect: true,
		records:    make(map[PhaseID]PhaseRecord),
	}
}

// CurrentPhase returns the active PhaseSpec.
func (m *Manager) CurrentPhase() PhaseSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	spec, ok := LookupPhase(string(m.current))
	if !ok {
		spec, _ = LookupPhase(string(PhaseImplementation))
	}
	return spec
}

// Methodology returns the active SDLC execution model ("agile", "waterfall", "devops").
func (m *Manager) Methodology() Methodology {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.method
}

// SetMethodology switches between Agile, Waterfall, and DevOps execution models.
func (m *Manager) SetMethodology(method string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch strings.ToLower(strings.TrimSpace(method)) {
	case string(MethodAgile), "scrum":
		m.method = MethodAgile
		return nil
	case string(MethodWaterfall), "linear":
		m.method = MethodWaterfall
		m.autoDetect = false // Waterfall enforces strict linear phase progression
		return nil
	case string(MethodDevOps), "cicd", "ci/cd":
		m.method = MethodDevOps
		return nil
	default:
		return fmt.Errorf("unknown SDLC methodology %q (choose agile, waterfall, or devops)", method)
	}
}

// AutoDetect reports whether Mímir automatically selects the SDLC phase from the prompt.
func (m *Manager) AutoDetect() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.autoDetect
}

// SetAutoDetect enables or disables automatic per-prompt SDLC phase routing.
func (m *Manager) SetAutoDetect(on bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoDetect = on
}

// SetPhase transitions to the requested SDLC phase. In Waterfall mode, jumping ahead
// requires the preceding phase to be signed off first.
func (m *Manager) SetPhase(idOrStep string) (PhaseSpec, error) {
	target, ok := LookupPhase(idOrStep)
	if !ok {
		return PhaseSpec{}, fmt.Errorf("unknown SDLC phase %q (use 1..7 or planning, analysis, design, implementation, testing, deployment, maintenance)", idOrStep)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	curSpec, _ := LookupPhase(string(m.current))
	if m.method == MethodWaterfall && target.Step > curSpec.Step {
		rec := m.records[m.current]
		if !rec.SignedOff {
			return curSpec, fmt.Errorf("waterfall gate blocked: phase %d (%s - %s) must be completed and signed off before advancing to phase %d (%s)",
				curSpec.Step, curSpec.Title, curSpec.Deliverable, target.Step, target.Title)
		}
	}
	m.current = target.ID
	return target, nil
}

// SaveDeliverable writes the primary deliverable markdown document for the current
// (or specified) phase into .mimir/sdlc/ and marks the phase signed off.
func (m *Manager) SaveDeliverable(phaseID PhaseID, content string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if phaseID == "" {
		phaseID = m.current
	}
	spec, ok := LookupPhase(string(phaseID))
	if !ok {
		return "", fmt.Errorf("unknown SDLC phase %q", phaseID)
	}
	fullPath := filepath.Join(m.workDir, filepath.FromSlash(spec.ArtifactFile))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	header := fmt.Sprintf("# SDLC Step %d: %s -- %s\n- **Methodology**: `%s`\n- **Persona**: `%s`\n- **Updated**: `%s`\n\n",
		spec.Step, spec.Title, spec.Deliverable, m.method, spec.Persona, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(fullPath, []byte(header+strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		return "", err
	}
	m.records[spec.ID] = PhaseRecord{
		Phase:       spec.ID,
		SignedOff:   true,
		Deliverable: spec.ArtifactFile,
		UpdatedAt:   time.Now(),
	}
	return spec.ArtifactFile, nil
}

// SignOff marks the active phase as signed off without overwriting its deliverable file.
func (m *Manager) SignOff(phaseID PhaseID) PhaseSpec {
	m.mu.Lock()
	defer m.mu.Unlock()
	if phaseID == "" {
		phaseID = m.current
	}
	spec, _ := LookupPhase(string(phaseID))
	rec := m.records[spec.ID]
	rec.Phase = spec.ID
	rec.SignedOff = true
	rec.UpdatedAt = time.Now()
	if rec.Deliverable == "" {
		rec.Deliverable = spec.ArtifactFile
	}
	m.records[spec.ID] = rec
	return spec
}

// Advance moves to the next sequential phase according to the active SDLC Methodology:
// - Agile: cycles 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 1 (iterative sprint loop)
// - Waterfall: strictly requires current phase SignedOff before moving to Step + 1
// - DevOps: chains Implementation (4) -> Testing (5) -> Deployment (6) -> Maintenance (7)
func (m *Manager) Advance() (PhaseSpec, error) {
	m.mu.Lock()
	curSpec, _ := LookupPhase(string(m.current))
	method := m.method
	rec := m.records[m.current]
	m.mu.Unlock()

	if method == MethodWaterfall && !rec.SignedOff {
		return curSpec, fmt.Errorf("waterfall gate: sign off Step %d (%s - %s) before advancing",
			curSpec.Step, curSpec.Title, curSpec.Deliverable)
	}

	nextStep := curSpec.Step + 1
	if method == MethodAgile && nextStep > 6 {
		nextStep = 1 // Agile sprint loop (1..6)
	} else if nextStep > 7 {
		nextStep = 7
	}
	return m.SetPhase(fmt.Sprintf("%d", nextStep))
}

// DetectPhaseFromPrompt classifies a user prompt into the matching SDLC PhaseID
// in <1ms so Mímir loads ONLY that phase's compact micro-prompt without bloating the base prompt.
func DetectPhaseFromPrompt(input string) PhaseID {
	q := strings.ToLower(strings.TrimSpace(input))
	if q == "" {
		return PhaseImplementation
	}

	// Step 1: Planning
	for _, kw := range []string{"project plan", "charter", "scope", "feasibility", "roadmap", "estimate cost", "sprint plan", "todo list"} {
		if strings.Contains(q, kw) {
			return PhasePlanning
		}
	}
	// Step 2: Analysis
	for _, kw := range []string{"requirements", "srs", "stakeholder", "constraints", "analyze repo", "ingest", "what does this", "explain how"} {
		if strings.Contains(q, kw) {
			return PhaseAnalysis
		}
	}
	// Step 3: Design
	for _, kw := range []string{"architecture", "blueprint", "database schema", "db schema", "wireframe", "sdd", "system design", "api contract", "folder structure"} {
		if strings.Contains(q, kw) {
			return PhaseDesign
		}
	}
	// Step 5: Testing
	for _, kw := range []string{"run test", "unit test", "qa report", "verify", "linter", "benchmark", "squash bug", "test suite"} {
		if strings.Contains(q, kw) {
			return PhaseTesting
		}
	}
	// Step 6: Deployment
	for _, kw := range []string{"deploy", "dockerfile", "ci/cd", "release", "production", "container", "hostinger", "dokploy"} {
		if strings.Contains(q, kw) {
			return PhaseDeployment
		}
	}
	// Step 7: Maintenance
	for _, kw := range []string{"monitor", "telemetry", "uptime", "patch vulnerability", "log audit", "maintenance", "hotfix"} {
		if strings.Contains(q, kw) {
			return PhaseMaintenance
		}
	}
	// Default Step 4: Implementation
	return PhaseImplementation
}

// ResolvePhaseForTurn returns the active PhaseSpec for a turn (auto-switching when
// AutoDetect is enabled and Methodology is not strict Waterfall).
func (m *Manager) ResolvePhaseForTurn(input string) PhaseSpec {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.autoDetect && m.method != MethodWaterfall && strings.TrimSpace(input) != "" {
		m.current = DetectPhaseFromPrompt(input)
	}
	spec, ok := LookupPhase(string(m.current))
	if !ok {
		spec, _ = LookupPhase(string(PhaseImplementation))
	}
	return spec
}

// BuildPhasePrompt returns ONLY the compact micro-prompt for the active SDLC step
// (plus the 1-line Web/SEO overlay ONLY if isWebTask is true and the phase has one,
// and compact PRODUCT.md / DESIGN.md context during Steps 3-5 when present).
// This keeps total lifecycle prompt overhead under ~75 tokens instead of bloating the system prompt.
func (m *Manager) BuildPhasePrompt(input string, isWebTask bool) string {
	spec := m.ResolvePhaseForTurn(input)
	m.mu.RLock()
	method := m.method
	workDir := m.workDir
	m.mu.RUnlock()

	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(spec.MicroPrompt)
	b.WriteString(fmt.Sprintf(" (Method: %s | Deliverable: %s)", method, spec.Deliverable))
	if isWebTask && spec.WebOverlay != "" {
		b.WriteString("\n")
		b.WriteString(spec.WebOverlay)
		if spec.Step >= 3 && spec.Step <= 5 {
			if dc := loadCompactDesignDocs(workDir); dc != "" {
				b.WriteString("\n")
				b.WriteString(dc)
			}
		}
	}
	return b.String()
}

func loadCompactDesignDocs(workDir string) string {
	if strings.TrimSpace(workDir) == "" {
		workDir = "."
	}
	var parts []string
	for _, name := range []string{"PRODUCT.md", "DESIGN.md"} {
		raw, err := os.ReadFile(filepath.Join(workDir, name))
		if err != nil {
			continue
		}
		compact := strings.Join(strings.Fields(strings.TrimSpace(string(raw))), " ")
		if compact == "" {
			continue
		}
		if len(compact) > 220 {
			compact = compact[:220] + "..."
		}
		parts = append(parts, fmt.Sprintf("[%s]: %s", name, compact))
	}
	return strings.Join(parts, " | ")
}

// AllowedToolSet returns a map of tool names permitted in the active phase.
func (m *Manager) AllowedToolSet() map[string]bool {
	spec := m.CurrentPhase()
	allowed := make(map[string]bool, len(spec.AllowedTools))
	for _, name := range spec.AllowedTools {
		allowed[name] = true
	}
	return allowed
}

// FormatBoard renders a clean markdown/terminal table of all 7 SDLC phases,
// highlighting the active phase and sign-off status.
func (m *Manager) FormatBoard() string {
	m.mu.RLock()
	cur := m.current
	method := m.method
	auto := m.autoDetect
	records := make(map[PhaseID]PhaseRecord, len(m.records))
	for k, v := range m.records {
		records[k] = v
	}
	m.mu.RUnlock()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("SDLC Framework [Methodology: %s | Auto-Phase: %v]\n", method, auto))
	for _, p := range AllPhases() {
		marker := "  "
		if p.ID == cur {
			marker = "->"
		}
		status := "pending"
		if rec, ok := records[p.ID]; ok && rec.SignedOff {
			status = "signed-off (" + rec.Deliverable + ")"
		} else if p.ID == cur {
			status = "ACTIVE"
		}
		b.WriteString(fmt.Sprintf("%s [%d/7] %-14s | Persona: %-8s | Deliverable: %-38s | %s\n",
			marker, p.Step, p.Title, p.Persona, p.Deliverable, status))
	}
	return strings.TrimRight(b.String(), "\n")
}
