package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditWebDirectory_CatchesImpeccableSlopAndMissingSEOGEO(t *testing.T) {
	dir := t.TempDir()

	slopHTML := "<!doctype html>\n<html><head></head><body>\n" +
		"<div class=\"card\"><div class=\"card bg-gradient-to-r from-purple-600 to-indigo-600\">\n" +
		"<p>In today's fast-paced world \u2014 we make your workflow 10x faster!</p>\n" +
		"</div></div>\n</body></html>"

	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(slopHTML), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	rep, err := AuditWebDirectory(dir)
	if err != nil {
		t.Fatalf("AuditWebDirectory: %v", err)
	}
	if rep.PassAudit {
		t.Fatalf("expected slop HTML to fail audit, got PASS:\n%s", rep.FormatMarkdown())
	}

	rulesHit := make(map[string]bool)
	for _, f := range rep.Findings {
		rulesHit[f.RuleID] = true
	}

	expectedRules := []string{
		"IMPECCABLE_EM_DASH",
		"IMPECCABLE_AI_SLOP_COPY",
		"IMPECCABLE_PURPLE_GRADIENT",
		"IMPECCABLE_NESTED_CARDS",
		"PROVENANCE_UNLABELED_STAT",
		"IMPECCABLE_HEADING_HIERARCHY",
		"SEO_MISSING_TITLE",
		"GEO_MISSING_SCHEMA_JSONLD",
		"GEO_MISSING_ROBOTS_TXT",
		"GEO_MISSING_SITEMAP_XML",
		"GEO_MISSING_LLMS_TXT",
	}
	for _, rule := range expectedRules {
		if !rulesHit[rule] {
			t.Errorf("expected rule %s to fire, findings=%+v", rule, rep.Findings)
		}
	}
}

func TestAuditWebDirectory_PassesCleanSemanticSiteWithDurableContext(t *testing.T) {
	dir := t.TempDir()

	cleanHTML := `<!doctype html>
<html lang="en">
<head>
  <title>PulseDock - Postgres Replica Lag Monitor ($29/mo)</title>
  <meta name="description" content="Real-time Postgres replication lag alerts with 60-second PagerDuty handoff.">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "name": "PulseDock",
    "applicationCategory": "DeveloperApplication"
  }
  </script>
</head>
<body>
  <header><nav><a href="/">PulseDock</a></nav></header>
  <main>
    <h1>Postgres Replication Lag Alerts in 5 Seconds</h1>
    <p>Pro plan starts at $29/mo per cluster. Reduces alert latency by 40% [vendor claim].</p>
  </main>
  <footer><p>Contact: ops@pulsedock.dev</p></footer>
</body>
</html>`

	files := map[string]string{
		"index.html":  cleanHTML,
		"robots.txt":  "User-agent: *\nAllow: /\nSitemap: https://pulsedock.dev/sitemap.xml\n",
		"sitemap.xml": "<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\"></urlset>\n",
		"llms.txt":    "# PulseDock\n> Postgres replica lag monitor ($29/mo per cluster).\n",
		"PRODUCT.md":  "# PulseDock\nTarget audience: Postgres DBAs and SRE teams.",
		"DESIGN.md":   "# Design System\nPalette: Obsidian #0E1116, Signal Amber #F59E0B. Font: Space Grotesk.",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	rep, err := AuditWebDirectory(dir)
	if err != nil {
		t.Fatalf("AuditWebDirectory: %v", err)
	}
	if !rep.PassAudit || len(rep.Findings) != 0 {
		t.Fatalf("expected clean semantic site to pass with 0 findings, got:\n%s", rep.FormatMarkdown())
	}
	if !rep.HasProductDoc || !rep.HasDesignDoc || rep.ValidJSONLDCount != 1 {
		t.Fatalf("expected durable context + 1 valid JSON-LD block, got %+v", rep)
	}

	ctxText := LoadDesignContext(dir)
	if !strings.Contains(ctxText, "[PRODUCT.md]:") || !strings.Contains(ctxText, "[DESIGN.md]:") {
		t.Fatalf("expected LoadDesignContext to include PRODUCT.md and DESIGN.md, got: %s", ctxText)
	}

	// Verify WebAuditTool.Run
	tool := &WebAuditTool{WorkDir: dir}
	out, err := tool.Run(context.Background(), map[string]any{"path": dir})
	if err != nil || !strings.Contains(out, "`PASS`") {
		t.Fatalf("WebAuditTool.Run failed: out=%s err=%v", out, err)
	}
}

// N9: the web audit tool confines its model-supplied path to the tool's
// workspace, like the file tools.
func TestWebAuditTool_RejectsPathOutsideWorkspace(t *testing.T) {
	wsDir := t.TempDir()
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := &WebAuditTool{WorkDir: wsDir}
	if _, err := tool.Run(context.Background(), map[string]any{"path": outsideDir}); err == nil {
		t.Fatal("expected security error auditing a directory outside the workspace")
	}
	if _, err := tool.Run(context.Background(), map[string]any{"path": ".."}); err == nil {
		t.Fatal("expected security error for relative escape")
	}
	// The workspace itself audits fine.
	if _, err := tool.Run(context.Background(), map[string]any{"path": "."}); err != nil {
		t.Fatalf("expected in-workspace audit to succeed: %v", err)
	}
}
