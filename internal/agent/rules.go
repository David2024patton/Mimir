package agent

import (
	"strings"

	"github.com/David2024patton/Mimir/internal/sdlc"
)

// CoreIdentityLine is the single-sentence identity header (<25 tokens) that keeps
// the base prompt razor-thin for both frontier and tiny local models.
const CoreIdentityLine = "Mímir is a modular polyglot Software Engineering & Creation Harness (Go + Rust + Mojo + Flutter). Never use em dashes."

// WebAndMarketingCreationRules holds the full 4-Layer SEO/GEO & Anti-Slop specification
// for on-demand auditing/skills (NOT injected in bulk into the base prompt).
const WebAndMarketingCreationRules = `
[MÍMIR 4-LAYER WEB, SEO/GEO & ANTI-SLOP REFERENCE SPEC]
1. TECHNICAL DISCOVERABILITY (SEO + GEO/AEO): Semantic HTML5, Core Web Vitals, JSON-LD Schema.org (LocalBusiness, SoftwareApplication, Product, FAQPage), /robots.txt, /sitemap.xml, and /llms.txt.
2. ENTITY CLARITY: Explicit machine-readable facts (NAP consistency for Google Maps/GBP, service areas, exact pricing units, direct-answer FAQ blocks).
3. EVIDENCE PROVENANCE & ANTI-SLOP: Zero generic AI hero clichés, zero fabricated statistics/reviews (FTC Review Rule), label vendor numbers as [vendor claim].
4. SPEED-TO-LEAD ARCHITECTURE: First-party lead record (source, response_time, contact, booked_outcome, next_action) + 1-click human handoff.`

var webCreationKeywords = []string{
	"website", "web app", "webapp", "landing page", "frontend", "html", "css",
	"nextjs", "react", "astro", "tailwind", "flutter", "seo", "geo", "aeo",
	"schema.org", "json-ld", "llms.txt", "sitemap", "google maps", "gbp",
	"local business", "marketing", "lead gen", "conversion", "saas", "pricing page",
	"blog", "hero section", "ui", "ux",
}

// DetectWebCreationTask returns true when the prompt involves building, designing,
// auditing, or optimizing a website, web/mobile app, UI, or SEO/GEO marketing surface.
func DetectWebCreationTask(input string) bool {
	lower := strings.ToLower(input)
	for _, kw := range webCreationKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// BuildPhaseScopedRulesBlock returns ONLY the active SDLC phase's compact micro-prompt
// (~40-75 tokens) + optional 1-line Web/SEO overlay for that specific phase, keeping
// the main prompt completely free of multi-phase bloat for both big and tiny LLMs.
func BuildPhaseScopedRulesBlock(mgr *sdlc.Manager, input string) string {
	if mgr == nil {
		mgr = sdlc.NewManager(".")
	}
	isWeb := DetectWebCreationTask(input)
	return "\n" + CoreIdentityLine + mgr.BuildPhasePrompt(input, isWeb)
}

// BuildHarnessRulesBlock is a convenience wrapper around BuildPhaseScopedRulesBlock.
func BuildHarnessRulesBlock(input string, _ bool) string {
	return BuildPhaseScopedRulesBlock(sdlc.NewManager("."), input)
}

// StopGateStatus summarizes the Level 1-3 Harness OnStop verification checks.
type StopGateStatus struct {
	HasPendingTodos bool
	PendingCount    int
	RanVerification bool
	Advisory        string
}

// EvaluateStopGate inspects the current turn's tool trace and Todo store to verify
// whether the agent completed its checklist and ran verification before stopping.
func EvaluateStopGate(trace []Step, todoLines []string) StopGateStatus {
	var pending int
	for _, line := range todoLines {
		if strings.Contains(line, "[pending]") || strings.Contains(line, "[in_progress]") {
			pending++
		}
	}
	ranVerify := false
	mutatedFiles := false
	for _, st := range trace {
		if st.Name == "write_file" || st.Name == "edit_block" {
			mutatedFiles = true
		}
		if st.Name == "run_shell" {
			lowerArgs := strings.ToLower(st.Args)
			if strings.Contains(lowerArgs, "test") || strings.Contains(lowerArgs, "vet") ||
				strings.Contains(lowerArgs, "lint") || strings.Contains(lowerArgs, "build") ||
				strings.Contains(lowerArgs, "check") {
				ranVerify = true
			}
		}
	}

	status := StopGateStatus{
		HasPendingTodos: pending > 0,
		PendingCount:    pending,
		RanVerification: !mutatedFiles || ranVerify,
	}

	var reasons []string
	if status.HasPendingTodos {
		reasons = append(reasons, "pending todo items remain")
	}
	if mutatedFiles && !ranVerify {
		reasons = append(reasons, "files were modified without running build/lint/test verification")
	}
	if len(reasons) > 0 {
		status.Advisory = "Harness OnStop Gate Advisory: " + strings.Join(reasons, "; ")
	}
	return status
}
