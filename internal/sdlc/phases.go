package sdlc

import (
	"fmt"
	"strings"
)

// PhaseID identifies one of the 7 sequential phases of the Software Development Life Cycle.
type PhaseID string

const (
	PhasePlanning       PhaseID = "planning"       // Step 1: Project Plan & Charter
	PhaseAnalysis       PhaseID = "analysis"       // Step 2: Software Requirement Specification (SRS)
	PhaseDesign         PhaseID = "design"         // Step 3: System Design Document (SDD)
	PhaseImplementation PhaseID = "implementation" // Step 4: Working Source Code & Builds
	PhaseTesting        PhaseID = "testing"        // Step 5: QA Test Reports & Bug-Free Build
	PhaseDeployment     PhaseID = "deployment"     // Step 6: Live Production Release
	PhaseMaintenance    PhaseID = "maintenance"    // Step 7: Updates, Patches & Logs
)

// Methodology defines how the harness moves through the 7 SDLC steps.
type Methodology string

const (
	MethodAgile     Methodology = "agile"     // Iterative sprints looping steps 1-6 continuously
	MethodWaterfall Methodology = "waterfall" // Linear: each phase must be 100% signed off before next begins
	MethodDevOps    Methodology = "devops"    // Continuous: automates Implementation (4) -> Testing (5) -> Deployment (6)
)

// PhaseSpec defines the metadata, deliverable, allowed tools, and compact micro-prompt
// for a single SDLC phase so the main prompt never suffers from instruction bloat.
type PhaseSpec struct {
	Step         int
	ID           PhaseID
	Title        string
	CoreGoal     string
	Activities   string
	Deliverable  string
	ArtifactFile string
	Persona      string
	AllowedTools []string
	MicroPrompt  string
	WebOverlay   string // Optional compact overlay loaded ONLY on web/app tasks during relevant phases
}

// AllPhases returns the 7 ordered SDLC phase specifications.
func AllPhases() []PhaseSpec {
	return []PhaseSpec{
		{
			Step:         1,
			ID:           PhasePlanning,
			Title:        "Planning",
			CoreGoal:     "Define project scope & feasibility",
			Activities:   "Set objectives, estimate costs/schedules, allocate resources, and evaluate technical risk.",
			Deliverable:  "Project Plan & Charter",
			ArtifactFile: ".mimir/sdlc/01-PLAN.md",
			Persona:      "odin",
			AllowedTools: []string{"read_file", "list_directory", "ingest_repo", "todo", "question", "sdlc"},
			MicroPrompt:  "[SDLC 1/7: PLANNING (Odin)] Define scope, feasibility, risks, and modular milestones. Record a step-by-step checklist in `todo` and write the Project Plan & Charter via `sdlc` before coding.",
		},
		{
			Step:         2,
			ID:           PhaseAnalysis,
			Title:        "Analysis",
			CoreGoal:     "Gather exact system requirements",
			Activities:   "Clarify requirements, define functional & security constraints, and map user workflows.",
			Deliverable:  "Software Requirement Specification (SRS)",
			ArtifactFile: ".mimir/sdlc/02-SRS.md",
			Persona:      "huginn",
			AllowedTools: []string{"read_file", "list_directory", "ingest_repo", "todo", "question", "sdlc"},
			MicroPrompt:  "[SDLC 2/7: ANALYSIS (Huginn)] Inspect the Yggdrasil Mind Map (`ingest_repo`/`read_file`), clarify ambiguities with `question`, and define exact functional/security constraints in the SRS.",
			WebOverlay:   "[Web/SEO Analysis]: Define target entity facts (NAP, pricing units, service areas) and first-party lead conversion fields (`source`, `response_time`, `contact`, `outcome`).",
		},
		{
			Step:         3,
			ID:           PhaseDesign,
			Title:        "Design",
			CoreGoal:     "Blueprint the system architecture",
			Activities:   "Design database schemas, API/FFI contracts, UI/UX wireframes, and modular folder structure.",
			Deliverable:  "System Design Document (SDD)",
			ArtifactFile: ".mimir/sdlc/03-SDD.md",
			Persona:      "odin",
			AllowedTools: []string{"read_file", "write_file", "list_directory", "todo", "question", "sdlc"},
			MicroPrompt:  "[SDLC 3/7: DESIGN (Odin)] Blueprint modular package/folder boundaries, database schemas, C-ABI/API contracts, and UI wireframes in the SDD before implementation.",
			WebOverlay:   "[Web/SEO Design]: Blueprint semantic HTML5 landmarks, JSON-LD Schema.org types (`LocalBusiness`/`SoftwareApplication`/`FAQPage`), `/llms.txt` structure, and non-formulaic UI layout.",
		},
		{
			Step:         4,
			ID:           PhaseImplementation,
			Title:        "Implementation",
			CoreGoal:     "Write the actual source code",
			Activities:   "Write modular backend/frontend code, keep edits surgical, and manage clean files.",
			Deliverable:  "Working Source Code & Builds",
			ArtifactFile: ".mimir/sdlc/04-BUILD.md",
			Persona:      "thor",
			AllowedTools: []string{"read_file", "write_file", "edit_block", "list_directory", "bash", "todo", "load_skill", "sdlc"},
			MicroPrompt:  "[SDLC 4/7: IMPLEMENTATION (Thor)] Read target files first, make surgical `edit_block` changes (<15 lines context), keep packages modular, and never use em dashes or placeholder stubs.",
			WebOverlay:   "[Web/SEO Build]: Emit semantic HTML5, JSON-LD Schema.org, `/llms.txt`, explicit pricing/entity facts, and first-party lead capture with human handoff. Zero AI slop or fake stats (`[vendor claim]`).",
		},
		{
			Step:         5,
			ID:           PhaseTesting,
			Title:        "Testing",
			CoreGoal:     "Verify quality & squash bugs",
			Activities:   "Run unit, integration, security, and linter checks against Step 2 requirements.",
			Deliverable:  "QA Test Reports & Bug-Free Build",
			ArtifactFile: ".mimir/sdlc/05-QA-REPORT.md",
			Persona:      "loki",
			AllowedTools: []string{"read_file", "write_file", "edit_block", "bash", "todo", "sdlc"},
			MicroPrompt:  "[SDLC 5/7: TESTING (Loki)] Run linter, static analysis, and affected unit/integration tests via `bash`. Fix every failure and record the QA Test Report via `sdlc`.",
			WebOverlay:   "[Web/SEO QA]: Verify valid JSON-LD Schema.org, `/robots.txt`, `/sitemap.xml`, `/llms.txt`, zero em dashes, and FTC-compliant factual claims.",
		},
		{
			Step:         6,
			ID:           PhaseDeployment,
			Title:        "Deployment",
			CoreGoal:     "Release software to production",
			Activities:   "Build release artifacts, verify containers/configs, and run release readiness checks.",
			Deliverable:  "Live Production Release",
			ArtifactFile: ".mimir/sdlc/06-RELEASE.md",
			Persona:      "heimdall",
			AllowedTools: []string{"read_file", "write_file", "edit_block", "bash", "todo", "sdlc"},
			MicroPrompt:  "[SDLC 6/7: DEPLOYMENT (Heimdall)] Verify release binaries/containers, environment configs, and rollback readiness. Never expose secrets. Record the Release manifest via `sdlc`.",
			WebOverlay:   "[Web/SEO Release Gate]: Confirm live canonical tags, `/llms.txt`, Search Console / GBP indexability, and active first-party lead beacon.",
		},
		{
			Step:         7,
			ID:           PhaseMaintenance,
			Title:        "Maintenance",
			CoreGoal:     "Monitor, patch, and iterate",
			Activities:   "Track telemetry/logs, patch vulnerabilities, resolve regressions, and queue next sprint.",
			Deliverable:  "Updates, Patches & Logs",
			ArtifactFile: ".mimir/sdlc/07-MAINTENANCE.md",
			Persona:      "muninn",
			AllowedTools: []string{"read_file", "edit_block", "bash", "ingest_repo", "todo", "sdlc"},
			MicroPrompt:  "[SDLC 7/7: MAINTENANCE (Muninn)] Inspect logs/telemetry, apply targeted regression patches, update the Yggdrasil Mind Map (`ingest_repo`), and queue new sprint items in `todo`.",
		},
	}
}

// LookupPhase returns the PhaseSpec for a given PhaseID or step number ("1".."7").
func LookupPhase(idOrStep string) (PhaseSpec, bool) {
	q := strings.ToLower(strings.TrimSpace(idOrStep))
	for _, p := range AllPhases() {
		if string(p.ID) == q || fmt.Sprintf("%d", p.Step) == q || strings.ToLower(p.Title) == q {
			return p, true
		}
	}
	return PhaseSpec{}, false
}
