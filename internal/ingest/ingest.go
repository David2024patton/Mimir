// Package ingest implements Mímir's built-in GitIngest engine ("Gjallarhorn-Ingest")
// and Yggdrasil Mind Map builder. It turns any local directory or Git repository URL
// into:
//  1. A GitIngest-formatted text digest (Directory structure tree + FILE blocks)
//  2. A structured Yggdrasil Mind Map (nodes, symbols, imports, and test links)
//  3. Persistent Cortex neurons so Mímir and the Liquid AI d1-omni-600M Decision
//     Model can query the repository's Mind Map across sessions.
package ingest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/David2024patton/Mimir/internal/cortex"
)

const (
	DefaultMaxFileSize = 32 * 1024 * 1024 // 32 MB per file so large GitIngest .txt dumps in INGEST/ are indexed
	DefaultMaxFiles    = 1500
)

// Options configures a repository or directory ingestion run.
type Options struct {
	Source        string // local path or https://github.com/owner/repo URL
	Branch        string
	MaxFileSize   int64
	MaxFiles      int
	WriteDigestTo string // optional file path to write digest.txt
}

// MindNode represents one file or module inside the Yggdrasil Mind Map.
type MindNode struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	Bytes    int64    `json:"bytes"`
	Symbols  []string `json:"symbols,omitempty"` // top exported functions/types/classes
	Imports  []string `json:"imports,omitempty"` // imported packages/modules
	IsTest   bool     `json:"is_test,omitempty"` // true for *_test.go, *.spec.ts, test_*.py
	Summary  string   `json:"summary,omitempty"` // first doc comment or header line
}

// MindMap is Mímir's structured structural graph ("Yggdrasil Mind Map") of a repository.
// It is stored in the Cortex and fed directly into the Liquid AI d1-omni-600M Decision Model.
type MindMap struct {
	RepoName   string     `json:"repo_name"`
	Source     string     `json:"source"`
	IngestedAt string     `json:"ingested_at"`
	TotalFiles int        `json:"total_files"`
	TotalBytes int64      `json:"total_bytes"`
	EstTokens  int        `json:"est_tokens"`
	Tree       string     `json:"tree"`
	Nodes      []MindNode `json:"nodes"`
}

// Result contains the full GitIngest output and Yggdrasil Mind Map.
type Result struct {
	Summary      string  `json:"summary"`
	Tree         string  `json:"tree"`
	Digest       string  `json:"digest"`
	MindMap      MindMap `json:"mind_map"`
	NeuronsSaved int     `json:"neurons_saved"`
}

var defaultIgnoredDirs = map[string]bool{
	".git": true, ".hg": true, ".svn": true, ".mimir": true,
	"node_modules": true, "vendor": true, "dist": true, "build": true,
	"target": true, "__pycache__": true, ".venv": true, "venv": true,
	".next": true, ".idea": true, ".vscode": true, "bin": true, "obj": true,
	"INGEST": true,
}

var binaryExts = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".a": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".ico": true, ".pdf": true, ".zip": true, ".tar": true, ".gz": true,
	".7z": true, ".mp3": true, ".mp4": true, ".wav": true, ".woff": true,
	".woff2": true, ".ttf": true, ".eot": true, ".pyc": true, ".db": true,
	".sqlite": true, ".docx": true, ".xlsx": true, ".pptx": true,
}

var (
	reGoSym   = regexp.MustCompile(`^(?:func|type)\s+(?:\([^)]+\)\s*)?([A-Za-z_][A-Za-z0-9_]*)`)
	rePySym   = regexp.MustCompile(`^(?:def|class|async\s+def)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	reTSSym   = regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?(?:function|class|interface|type|const)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	reRustSym = regexp.MustCompile(`^(?:pub\s+)?(?:async\s+)?(?:fn|struct|enum|trait|mod)\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

// Run ingests a local path or remote Git URL (GitIngest-compatible), builds the
// Yggdrasil Mind Map, optionally writes a digest file, and stores the Mind Map +
// module digests into Cortex memory if store != nil.
func Run(ctx context.Context, opts Options, store cortex.Store) (*Result, error) {
	if opts.MaxFileSize <= 0 {
		opts.MaxFileSize = DefaultMaxFileSize
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = DefaultMaxFiles
	}
	source := strings.TrimSpace(opts.Source)
	if source == "" {
		source = "."
	}

	rootDir := source
	var repoName string
	if isGitURL(source) {
		tmpDir, cleanup, err := shallowClone(ctx, source, opts.Branch)
		if err != nil {
			return nil, err
		}
		defer cleanup()
		rootDir = tmpDir
		repoName = repoNameFromURL(source)
	} else {
		abs, err := filepath.Abs(source)
		if err != nil {
			return nil, err
		}
		rootDir = abs
		repoName = filepath.Base(abs)
	}

	var relPaths []string
	var nodes []MindNode
	var filesBuf strings.Builder
	var totalBytes int64

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if path != rootDir && (defaultIgnoredDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if len(relPaths) >= opts.MaxFiles {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if binaryExts[ext] || strings.HasSuffix(d.Name(), ".lock") {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() == 0 || info.Size() > opts.MaxFileSize {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || isBinaryContent(data) {
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			rel = d.Name()
		}
		rel = filepath.ToSlash(rel)
		relPaths = append(relPaths, rel)
		totalBytes += info.Size()

		content := string(data)
		node := extractMindNode(rel, content, info.Size())
		nodes = append(nodes, node)

		filesBuf.WriteString("================================================\n")
		filesBuf.WriteString("FILE: " + rel + "\n")
		filesBuf.WriteString("================================================\n")
		filesBuf.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			filesBuf.WriteString("\n")
		}
		filesBuf.WriteString("\n")
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(relPaths)
	treeStr := renderTree(repoName, relPaths)
	estTokens := int(totalBytes / 4)

	summary := fmt.Sprintf(
		"Repository: %s\nFiles analyzed: %d\nEstimated tokens: ~%dk (%d bytes)",
		repoName, len(relPaths), estTokens/1000, totalBytes,
	)

	digest := "Directory structure:\n" + treeStr + "\n" + filesBuf.String()
	if opts.WriteDigestTo != "" {
		_ = os.WriteFile(opts.WriteDigestTo, []byte(digest), 0o644)
	}

	mm := MindMap{
		RepoName:   repoName,
		Source:     source,
		IngestedAt: time.Now().UTC().Format(time.RFC3339),
		TotalFiles: len(relPaths),
		TotalBytes: totalBytes,
		EstTokens:  estTokens,
		Tree:       treeStr,
		Nodes:      nodes,
	}

	saved := 0
	if store != nil {
		saved = persistToCortex(ctx, store, mm)
	}

	return &Result{
		Summary:      summary,
		Tree:         treeStr,
		Digest:       digest,
		MindMap:      mm,
		NeuronsSaved: saved,
	}, nil
}

// FormatMindMapCompact produces a dense, LLM- and Decision-Model-friendly view of
// the Yggdrasil Mind Map (directory tree + file symbols + test map).
func FormatMindMapCompact(mm MindMap, maxNodes int) string {
	if maxNodes <= 0 {
		maxNodes = 60
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Yggdrasil Mind Map: %s (%d files, ~%dk tokens)\n", mm.RepoName, mm.TotalFiles, mm.EstTokens/1000)
	b.WriteString("Directory Tree:\n")
	b.WriteString(mm.Tree)
	b.WriteString("\nSymbol & Dependency Graph:\n")
	for i, n := range mm.Nodes {
		if i >= maxNodes {
			fmt.Fprintf(&b, "  ... and %d more files\n", len(mm.Nodes)-maxNodes)
			break
		}
		tag := n.Language
		if n.IsTest {
			tag += ",test"
		}
		line := fmt.Sprintf("  • %s [%s]", n.Path, tag)
		if len(n.Symbols) > 0 {
			line += " symbols={" + strings.Join(n.Symbols, ", ") + "}"
		}
		if n.Summary != "" {
			line += " // " + n.Summary
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func persistToCortex(ctx context.Context, store cortex.Store, mm MindMap) int {
	saved := 0
	// 1. Save the master Yggdrasil Mind Map overview neuron
	compact := FormatMindMapCompact(mm, 80)
	rootID, err := store.PutNeuron(ctx, cortex.Neuron{
		Kind:    "mindmap",
		Layer:   "context",
		Title:   "Yggdrasil Mind Map: " + mm.RepoName,
		Content: compact,
	})
	if err == nil {
		saved++
	}
	// 2. Save raw JSON mindmap for programmatic Decision Model lookup
	if rawJSON, err := json.Marshal(mm); err == nil {
		if _, err := store.PutNeuron(ctx, cortex.Neuron{
			Kind:    "mindmap_json",
			Layer:   "context",
			Title:   "MindMap JSON: " + mm.RepoName,
			Content: string(rawJSON),
		}); err == nil {
			saved++
		}
	}
	// 3. Group nodes by top-level directory/module and store module-level knowledge neurons + synapses
	byDir := map[string][]MindNode{}
	for _, n := range mm.Nodes {
		dir := filepath.ToSlash(filepath.Dir(n.Path))
		byDir[dir] = append(byDir[dir], n)
	}
	for dir, dirNodes := range byDir {
		var sb strings.Builder
		fmt.Fprintf(&sb, "[Repo %s · Module %s] ", mm.RepoName, dir)
		for _, dn := range dirNodes {
			fmt.Fprintf(&sb, "%s", filepath.Base(dn.Path))
			if len(dn.Symbols) > 0 {
				fmt.Fprintf(&sb, "(%s)", strings.Join(dn.Symbols, ","))
			}
			sb.WriteString("; ")
		}
		modID, err := store.PutNeuron(ctx, cortex.Neuron{
			Kind:    cortex.KindSource,
			Layer:   "context",
			Title:   mm.RepoName + "/" + dir,
			Content: sb.String(),
		})
		if err == nil {
			saved++
			if rootID != "" {
				_ = store.Relate(ctx, cortex.Synapse{
					From: rootID,
					To:   modID,
					Kind: "references",
				})
			}
		}
	}
	// 4. When ingesting a flat reference corpus like INGEST/ (1 dir, multiple .txt/.md dumps),
	// also persist a dedicated searchable neuron per ingested reference file.
	if len(byDir) <= 2 && len(mm.Nodes) > 3 {
		for _, n := range mm.Nodes {
			if n.Language != "txt" && n.Language != "md" {
				continue
			}
			desc := fmt.Sprintf("[Ingested Corpus %s/%s (%d KB)] %s", mm.RepoName, n.Path, n.Bytes/1024, n.Summary)
			if len(n.Symbols) > 0 {
				desc += " | Key items: " + strings.Join(n.Symbols, ", ")
			}
			nID, err := store.PutNeuron(ctx, cortex.Neuron{
				Kind:    cortex.KindSource,
				Layer:   "context",
				Title:   mm.RepoName + "/" + n.Path,
				Content: desc,
			})
			if err == nil {
				saved++
				if rootID != "" {
					_ = store.Relate(ctx, cortex.Synapse{
						From: rootID,
						To:   nID,
						Kind: "references",
					})
				}
			}
		}
	}
	return saved
}

// LatestMindMap retrieves the most recently ingested Yggdrasil MindMap from Cortex memory.
func LatestMindMap(store cortex.Store) (*MindMap, bool) {
	if store == nil {
		return nil, false
	}
	all := store.All()
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Kind == "mindmap_json" {
			var mm MindMap
			if err := json.Unmarshal([]byte(all[i].Content), &mm); err == nil && mm.RepoName != "" {
				return &mm, true
			}
		}
	}
	return nil, false
}

func extractMindNode(relPath, content string, size int64) MindNode {
	ext := strings.ToLower(filepath.Ext(relPath))
	lang := strings.TrimPrefix(ext, ".")
	if lang == "" {
		lang = "text"
	}
	lowerPath := strings.ToLower(relPath)
	isTest := strings.Contains(lowerPath, "_test.") ||
		strings.Contains(lowerPath, ".test.") ||
		strings.Contains(lowerPath, ".spec.") ||
		strings.HasPrefix(filepath.Base(lowerPath), "test_")

	var symbols []string
	var imports []string
	summary := ""

	sc := bufio.NewScanner(strings.NewReader(content))
	sc.Buffer(make([]byte, 64*1024), 2*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if summary == "" && (strings.HasPrefix(line, "//") || strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "Repository:")) {
			clean := strings.TrimSpace(strings.TrimLeft(line, "/# "))
			if len(clean) > 6 && len(clean) < 120 {
				summary = clean
			}
		}
		if len(symbols) < 12 {
			if strings.HasPrefix(line, "FILE: ") {
				symbols = append(symbols, strings.TrimSpace(strings.TrimPrefix(line, "FILE: ")))
			} else if strings.HasPrefix(line, "## ") {
				symbols = append(symbols, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			} else {
				for _, re := range []*regexp.Regexp{reGoSym, rePySym, reTSSym, reRustSym} {
					if m := re.FindStringSubmatch(line); len(m) > 1 {
						symbols = append(symbols, m[1])
						break
					}
				}
			}
		}
		if len(imports) < 8 && (strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") || strings.HasPrefix(line, "use ")) {
			imports = append(imports, line)
		}
	}

	return MindNode{
		Path:     relPath,
		Language: lang,
		Bytes:    size,
		Symbols:  symbols,
		Imports:  imports,
		IsTest:   isTest,
		Summary:  summary,
	}
}

func renderTree(repoName string, relPaths []string) string {
	var b strings.Builder
	b.WriteString("└── " + repoName + "/\n")
	limit := len(relPaths)
	if limit > 120 {
		limit = 120
	}
	for i := 0; i < limit; i++ {
		prefix := "    ├── "
		if i == limit-1 && len(relPaths) <= limit {
			prefix = "    └── "
		}
		b.WriteString(prefix + relPaths[i] + "\n")
	}
	if len(relPaths) > limit {
		fmt.Fprintf(&b, "    └── ... (%d more files)\n", len(relPaths)-limit)
	}
	return b.String()
}

func isBinaryContent(data []byte) bool {
	check := len(data)
	if check > 1024 {
		check = 1024
	}
	for i := 0; i < check; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func isGitURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "git@")
}

func repoNameFromURL(u string) string {
	u = strings.TrimRight(u, "/")
	u = strings.TrimSuffix(u, ".git")
	parts := strings.Split(u, "/")
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		return parts[len(parts)-1]
	}
	return "repo"
}

func shallowClone(ctx context.Context, url, branch string) (string, func(), error) {
	// Sanitize url and branch to prevent git argument injection (F9)
	if strings.HasPrefix(url, "-") {
		return "", func() {}, fmt.Errorf("invalid repository url %q: cannot start with dash", url)
	}
	if strings.HasPrefix(branch, "-") {
		return "", func() {}, fmt.Errorf("invalid branch %q: cannot start with dash", branch)
	}

	tmp, err := os.MkdirTemp("", "mimir-gitingest-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }

	args := []string{"clone", "--depth=1", "--single-branch"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	// Use explicit "--" separator to prevent url from being parsed as options
	args = append(args, "--", url, tmp)

	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, "git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("git clone %s: %v (%s)", url, err, strings.TrimSpace(string(out)))
	}
	return tmp, cleanup, nil
}
