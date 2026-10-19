package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// WebFinding represents a single deterministic violation from the Impeccable Design
// or 4-Layer SEO/GEO & Conversion detector suite.
type WebFinding struct {
	RuleID      string `json:"rule_id"`
	Layer       string `json:"layer"`    // "impeccable_ui", "layer1_seo_geo", "layer2_entity", "layer3_provenance", "layer4_lead_gen"
	Severity    string `json:"severity"` // "HIGH", "MEDIUM", "LOW"
	File        string `json:"file"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

// WebAuditReport is the structured output of AuditWebDirectory.
type WebAuditReport struct {
	RootDir          string       `json:"root_dir"`
	ScannedFiles     int          `json:"scanned_files"`
	HasProductDoc    bool         `json:"has_product_md"`
	HasDesignDoc     bool         `json:"has_design_md"`
	HasRobotsTxt     bool         `json:"has_robots_txt"`
	HasSitemapXML    bool         `json:"has_sitemap_xml"`
	HasLLMsTxt       bool         `json:"has_llms_txt"`
	ValidJSONLDCount int          `json:"valid_jsonld_count"`
	Findings         []WebFinding `json:"findings"`
	PassAudit        bool         `json:"pass_audit"`
}

// FormatMarkdown renders the deterministic Web/SEO/GEO + Impeccable Design audit report.
func (r WebAuditReport) FormatMarkdown() string {
	var b strings.Builder
	verdict := "PASS"
	if !r.PassAudit {
		verdict = "FAIL"
	}
	b.WriteString(fmt.Sprintf("## Mimir Web, SEO/GEO & Impeccable Design Audit: `%s`\n", verdict))
	b.WriteString(fmt.Sprintf("- **Scanned Files**: `%d`\n", r.ScannedFiles))
	b.WriteString(fmt.Sprintf("- **Durable Context (`PRODUCT.md` / `DESIGN.md`)**: `product=%v`, `design=%v`\n", r.HasProductDoc, r.HasDesignDoc))
	b.WriteString(fmt.Sprintf("- **Discovery Files**: `robots.txt=%v`, `sitemap.xml=%v`, `llms.txt=%v`\n", r.HasRobotsTxt, r.HasSitemapXML, r.HasLLMsTxt))
	b.WriteString(fmt.Sprintf("- **Valid Schema.org JSON-LD Blocks**: `%d`\n", r.ValidJSONLDCount))
	b.WriteString(fmt.Sprintf("- **Total Findings**: `%d`\n", len(r.Findings)))

	if len(r.Findings) > 0 {
		b.WriteString("\n### Findings\n")
		for _, f := range r.Findings {
			b.WriteString(fmt.Sprintf("- **[%s | %s]** `%s` (%s): %s -- *Fix: %s*\n",
				f.Severity, f.RuleID, f.File, f.Layer, f.Description, f.Remediation))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

var (
	aiSlopPhrases = []string{
		"in today's fast-paced world",
		"unleash the power",
		"revolutionize the way",
		"delve into",
		"seamless digital landscape",
		"elevate your digital journey",
		"unlock the full potential",
		"game-changing solution",
		"synergistic ecosystem",
	}

	purpleGradientTokens = []string{
		"from-purple-",
		"from-indigo-",
		"via-purple-",
		"to-purple-",
		"to-indigo-",
		"linear-gradient(135deg, #8b5cf6",
		"linear-gradient(90deg, #6366f1",
	}

	reH1Tag          = regexp.MustCompile(`(?i)<h1[\s>]`)
	reJSONLD         = regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	reMarketingStats = regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?%\s+(?:faster|higher|more|increase|conversion|uptime)|\d+(?:\.\d+)?x\s+(?:faster|more|roi|growth))\b`)
	reNestedCard     = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*\bcard\b[^"']*["'][^>]*>.*?<div[^>]*class=["'][^"']*\bcard\b[^"']*["']`)
)

// LoadDesignContext reads durable PRODUCT.md and DESIGN.md files from workDir (if present)
// and returns a compact summary (<600 chars) for injection during SDLC Steps 3-5 on web/UI tasks.
func LoadDesignContext(workDir string) string {
	if strings.TrimSpace(workDir) == "" {
		workDir = "."
	}
	var parts []string
	for _, name := range []string{"PRODUCT.md", "DESIGN.md"} {
		p := filepath.Join(workDir, name)
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(raw))
		if text == "" {
			continue
		}
		// Collapse whitespace to keep context compact for both big and tiny models.
		compact := strings.Join(strings.Fields(text), " ")
		if len(compact) > 260 {
			compact = compact[:260] + "..."
		}
		parts = append(parts, fmt.Sprintf("[%s]: %s", name, compact))
	}
	return strings.Join(parts, "\n")
}

// AuditWebDirectory runs deterministic Impeccable UI anti-slop detectors and
// Mimir's 4-Layer SEO/GEO & Conversion detectors across a web project directory.
func AuditWebDirectory(rootDir string) (WebAuditReport, error) {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "."
	}
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return WebAuditReport{}, err
	}

	rep := WebAuditReport{
		RootDir:  absRoot,
		Findings: []WebFinding{},
	}

	if _, err := os.Stat(filepath.Join(absRoot, "PRODUCT.md")); err == nil {
		rep.HasProductDoc = true
	}
	if _, err := os.Stat(filepath.Join(absRoot, "DESIGN.md")); err == nil {
		rep.HasDesignDoc = true
	}

	var htmlFiles []string
	_ = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".next" || name == "dist" || name == ".mimir" {
				return filepath.SkipDir
			}
			return nil
		}
		lowerName := strings.ToLower(name)
		switch lowerName {
		case "robots.txt":
			rep.HasRobotsTxt = true
		case "sitemap.xml":
			rep.HasSitemapXML = true
		case "llms.txt":
			rep.HasLLMsTxt = true
		}

		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".html", ".htm", ".astro", ".tsx", ".jsx", ".vue", ".svelte", ".php", ".css":
			rep.ScannedFiles++
			rel, relErr := filepath.Rel(absRoot, path)
			if relErr != nil {
				rel = name
			}
			rel = filepath.ToSlash(rel)
			raw, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			content := string(raw)
			auditSingleWebFile(&rep, rel, ext, content)
			if ext == ".html" || ext == ".htm" || ext == ".astro" || ext == ".php" {
				htmlFiles = append(htmlFiles, rel)
			}
		}
		return nil
	})

	// Layer 1 Discovery file checks when at least one HTML/page file exists in the project.
	if len(htmlFiles) > 0 {
		if !rep.HasRobotsTxt {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "GEO_MISSING_ROBOTS_TXT",
				Layer:       "layer1_seo_geo",
				Severity:    "MEDIUM",
				File:        "robots.txt",
				Description: "Missing /robots.txt crawler directives file",
				Remediation: "Create robots.txt allowing search and AI answer engine crawlers and referencing Sitemap.",
			})
		}
		if !rep.HasSitemapXML {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "GEO_MISSING_SITEMAP_XML",
				Layer:       "layer1_seo_geo",
				Severity:    "MEDIUM",
				File:        "sitemap.xml",
				Description: "Missing /sitemap.xml index file",
				Remediation: "Generate sitemap.xml listing canonical URLs with <lastmod> timestamps.",
			})
		}
		if !rep.HasLLMsTxt {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "GEO_MISSING_LLMS_TXT",
				Layer:       "layer1_seo_geo",
				Severity:    "MEDIUM",
				File:        "llms.txt",
				Description: "Missing /llms.txt AI answer-engine summary file",
				Remediation: "Add /llms.txt with concise entity facts, product/service summary, pricing units, and canonical docs links.",
			})
		}
	}

	highCount := 0
	for _, f := range rep.Findings {
		if f.Severity == "HIGH" {
			highCount++
		}
	}
	rep.PassAudit = highCount == 0
	return rep, nil
}

func auditSingleWebFile(rep *WebAuditReport, relPath, ext, content string) {
	lower := strings.ToLower(content)

	// 1. Impeccable Detector: Em dash check (\u2014)
	if strings.Contains(content, "\u2014") || strings.Contains(lower, "&mdash;") {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "IMPECCABLE_EM_DASH",
			Layer:       "impeccable_ui",
			Severity:    "HIGH",
			File:        relPath,
			Description: "Contains banned em dash character or &mdash; entity",
			Remediation: "Replace em dashes with standard hyphens (-), double hyphens (--), colons, or clean sentence breaks.",
		})
	}

	// 2. Impeccable Detector: Formulaic AI Slop Copy
	for _, phrase := range aiSlopPhrases {
		if strings.Contains(lower, phrase) {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "IMPECCABLE_AI_SLOP_COPY",
				Layer:       "impeccable_ui",
				Severity:    "HIGH",
				File:        relPath,
				Description: fmt.Sprintf("Contains generic AI slop filler copy (%q)", phrase),
				Remediation: "Replace filler copy with specific, verifiable entity facts, concrete numbers, or direct product capabilities.",
			})
			break
		}
	}

	// 3. Impeccable Detector: Purple/Indigo Gradient Cliche
	for _, tok := range purpleGradientTokens {
		if strings.Contains(lower, tok) {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "IMPECCABLE_PURPLE_GRADIENT",
				Layer:       "impeccable_ui",
				Severity:    "MEDIUM",
				File:        relPath,
				Description: fmt.Sprintf("Uses formulaic AI purple/indigo gradient token (%q)", tok),
				Remediation: "Use intentional brand palette tokens defined in DESIGN.md instead of default purple/indigo gradients.",
			})
			break
		}
	}

	// 4. Impeccable Detector: Card-in-Card Nesting Soup
	if reNestedCard.MatchString(content) {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "IMPECCABLE_NESTED_CARDS",
			Layer:       "impeccable_ui",
			Severity:    "MEDIUM",
			File:        relPath,
			Description: "Nested .card inside .card container detected (visual clutter)",
			Remediation: "Flatten nested cards using clean spacing, dividers, or definition lists.",
		})
	}

	// 5. Layer 3 Provenance Detector: Unlabeled marketing stats in web copy
	if ext != ".css" {
		if matches := reMarketingStats.FindAllString(content, -1); len(matches) > 0 {
			if !strings.Contains(lower, "[vendor claim]") && !strings.Contains(lower, "source:") && !strings.Contains(lower, "citation") {
				rep.Findings = append(rep.Findings, WebFinding{
					RuleID:      "PROVENANCE_UNLABELED_STAT",
					Layer:       "layer3_provenance",
					Severity:    "HIGH",
					File:        relPath,
					Description: fmt.Sprintf("Unbacked performance/conversion claim %q without [vendor claim] or source citation", matches[0]),
					Remediation: "Ground the statistic in a cited benchmark or label first-party figures explicitly with [vendor claim].",
				})
			}
		}
	}

	// Full document HTML checks (only when file is a full HTML document with <html or <!doctype)
	isFullHTMLDoc := (ext == ".html" || ext == ".htm" || ext == ".php") &&
		(strings.Contains(lower, "<!doctype html") || strings.Contains(lower, "<html"))
	if !isFullHTMLDoc {
		return
	}

	// 6. Impeccable Heading Hierarchy: Exactly one <h1> per document
	h1Count := len(reH1Tag.FindAllStringIndex(content, -1))
	if h1Count == 0 {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "IMPECCABLE_HEADING_HIERARCHY",
			Layer:       "impeccable_ui",
			Severity:    "HIGH",
			File:        relPath,
			Description: "HTML document is missing a primary <h1> heading",
			Remediation: "Include exactly one descriptive <h1> heading stating the primary page entity and offer.",
		})
	} else if h1Count > 1 {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "IMPECCABLE_HEADING_HIERARCHY",
			Layer:       "impeccable_ui",
			Severity:    "MEDIUM",
			File:        relPath,
			Description: fmt.Sprintf("HTML document has %d <h1> tags (expected exactly 1)", h1Count),
			Remediation: "Keep a single primary <h1> and demote secondary section headers to <h2>/<h3>.",
		})
	}

	// 7. Layer 1 Technical SEO: <title>, <meta name="description">, semantic landmarks
	if !strings.Contains(lower, "<title>") {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "SEO_MISSING_TITLE",
			Layer:       "layer1_seo_geo",
			Severity:    "HIGH",
			File:        relPath,
			Description: "Missing <title> tag in HTML document",
			Remediation: "Add a unique <title> tag (50-60 chars) in <head>.",
		})
	}
	if !strings.Contains(lower, `name="description"`) && !strings.Contains(lower, `name='description'`) {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "SEO_MISSING_META_DESCRIPTION",
			Layer:       "layer1_seo_geo",
			Severity:    "MEDIUM",
			File:        relPath,
			Description: "Missing <meta name=\"description\"> tag in HTML document",
			Remediation: "Add a concise <meta name=\"description\" content=\"...\"> summarizing the page offer.",
		})
	}
	for _, landmark := range []string{"<header", "<main", "<footer"} {
		if !strings.Contains(lower, landmark) {
			rep.Findings = append(rep.Findings, WebFinding{
				RuleID:      "SEO_SEMANTIC_LANDMARKS",
				Layer:       "layer1_seo_geo",
				Severity:    "MEDIUM",
				File:        relPath,
				Description: fmt.Sprintf("Missing semantic HTML5 landmark %s>", landmark),
				Remediation: "Wrap page regions in semantic <header>, <nav>, <main>, and <footer> landmarks.",
			})
			break
		}
	}

	// 8. Layer 1 GEO/AEO: Schema.org JSON-LD validation
	scripts := reJSONLD.FindAllStringSubmatch(content, -1)
	if len(scripts) == 0 {
		rep.Findings = append(rep.Findings, WebFinding{
			RuleID:      "GEO_MISSING_SCHEMA_JSONLD",
			Layer:       "layer1_seo_geo",
			Severity:    "HIGH",
			File:        relPath,
			Description: "Missing <script type=\"application/ld+json\"> Schema.org structured data",
			Remediation: "Embed valid Schema.org JSON-LD (SoftwareApplication, LocalBusiness, Product, or FAQPage).",
		})
	} else {
		for _, match := range scripts {
			rawJSON := strings.TrimSpace(match[1])
			var parsed map[string]any
			if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
				rep.Findings = append(rep.Findings, WebFinding{
					RuleID:      "GEO_INVALID_SCHEMA_JSONLD",
					Layer:       "layer1_seo_geo",
					Severity:    "HIGH",
					File:        relPath,
					Description: fmt.Sprintf("Malformed JSON-LD structured data: %v", err),
					Remediation: "Fix syntax errors inside <script type=\"application/ld+json\"> so crawlers can parse it.",
				})
				continue
			}
			_, hasCtx := parsed["@context"]
			_, hasType := parsed["@type"]
			if !hasCtx || !hasType {
				rep.Findings = append(rep.Findings, WebFinding{
					RuleID:      "GEO_INCOMPLETE_SCHEMA_JSONLD",
					Layer:       "layer1_seo_geo",
					Severity:    "MEDIUM",
					File:        relPath,
					Description: "JSON-LD block is missing required @context or @type property",
					Remediation: "Include \"@context\": \"https://schema.org\" and an explicit \"@type\".",
				})
				continue
			}
			rep.ValidJSONLDCount++
		}
	}
}

// WebAuditTool exposes AuditWebDirectory as a native Mimir tool ("web_seo_audit").
type WebAuditTool struct {
	WorkDir string
}

func (t *WebAuditTool) Name() string { return "web_seo_audit" }

func (t *WebAuditTool) Description() string {
	return "Run deterministic Impeccable UI anti-slop detectors and 4-Layer SEO/GEO & Conversion checks (semantic HTML5, JSON-LD Schema.org, robots.txt, sitemap.xml, llms.txt, em dashes, purple gradient cliches, and unbacked claims) across a web directory."
}

func (t *WebAuditTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Target directory to audit (defaults to current workspace directory).",
			},
		},
	}
}

func (t *WebAuditTool) Run(_ context.Context, args map[string]any) (string, error) {
	pathArg, _ := args["path"].(string)
	target := strings.TrimSpace(pathArg)
	if target == "" {
		target = t.WorkDir
	}
	if target == "" {
		target = "."
	}
	// The path is model-supplied, so it gets the same workspace confinement
	// (including symlink resolution) as the file tools (N9).
	workDir := t.WorkDir
	if workDir == "" {
		workDir = "."
	}
	resolved, err := resolveWithinWorkDir(workDir, target)
	if err != nil {
		return "", fmt.Errorf("web_seo_audit: %w", err)
	}
	rep, err := AuditWebDirectory(resolved)
	if err != nil {
		return "", err
	}
	return rep.FormatMarkdown(), nil
}
