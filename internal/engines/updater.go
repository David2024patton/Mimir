package engines

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultUpdateBase is where `mimir update` pulls the published manifest and
// engine binaries from. Override with MIMIR_UPDATE_BASE (for forks, mirrors, or tests).
const DefaultUpdateBase = "https://raw.githubusercontent.com/David2024patton/Mimir/main"

// KnownEngines is the static registry of engine metadata. Version and SHA-256 are
// filled in at publish time from the actual compiled binary, never typed by hand.
var KnownEngines = []EngineEntry{
	{
		ID: "tyr-core", Name: "Tyr-Core Engine", Branch: "engine/tyr-rust",
		Binary: "engines/bin/mimir_tyr_engine.exe", Stack: "go (rust port planned)",
		Description: "Symbol outline and Okapi BM25 code search.",
	},
	{
		ID: "mjolnir-compute", Name: "Mjolnir-Compute Engine", Branch: "engine/mjolnir-mojo",
		Binary: "engines/bin/mimir_mjolnir_engine.exe", Stack: "go (mojo port planned)",
		Description: ".itf header I/O, cosine similarity, 10,048-bit VSA bind/unbind.",
	},
	{
		ID: "alphaflux-core", Name: "Alpha Flux Intelligence Engine", Branch: "engine/alphaflux-core",
		Binary: "engines/bin/mimir_alphaflux_engine.exe", Stack: "go",
		Description: "Huginn-Decider routing, Epistemic Critic, Impeccable + SEO/GEO web audit.",
	},
}

// UpdateBase returns the configured update source.
func UpdateBase() string {
	if v := strings.TrimSpace(os.Getenv("MIMIR_UPDATE_BASE")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultUpdateBase
}

// probeVersion runs `<binary> --attribution=... version` and reads the "version" field.
func probeVersion(ctx context.Context, binPath string) string {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, binPath, "--attribution="+AttributionHandshake, "version").Output()
	if err != nil {
		return "unknown"
	}
	var v struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(out, &v) != nil || v.Version == "" {
		return "unknown"
	}
	return v.Version
}

// BuildManifest hashes every known engine binary present in `engines/bin/`, asks each
// binary for its version, and writes `engines/bin/manifest.json`. This is the publish
// step: run it after copying freshly built binaries in, then commit to main.
func BuildManifest(ctx context.Context, rootDir string) (Manifest, error) {
	m := Manifest{
		SchemaVersion: "1.0",
		Owner:         RequiredOwner,
		Website:       "https://" + RequiredDomain,
		License:       "LICENSE-ENGINES.md",
		Attribution:   AttributionNotice,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	platforms := []struct {
		Key string // e.g. "windows_amd64", "linux_amd64", "darwin_arm64"
		Ext string // e.g. ".exe", "", ""
	}{
		{Key: "windows_amd64", Ext: ".exe"},
		{Key: "linux_amd64", Ext: ""},
		{Key: "darwin_arm64", Ext: ""},
	}

	for _, e := range KnownEngines {
		// Initialize Platforms map
		e.Platforms = make(map[string]PlatformBinary)

		// 1. Check for platform-specific binaries: e.g. engines/bin/mimir_tyr_engine_linux_amd64
		baseBinaryName := strings.TrimSuffix(filepath.Base(e.Binary), filepath.Ext(e.Binary))
		for _, p := range platforms {
			pBinRel := fmt.Sprintf("engines/bin/%s_%s%s", baseBinaryName, p.Key, p.Ext)
			pBinPath := filepath.Join(rootDir, filepath.FromSlash(pBinRel))
			if sum, size, err := FileSHA256(pBinPath); err == nil {
				e.Platforms[p.Key] = PlatformBinary{
					Binary:    pBinRel,
					SHA256:    sum,
					SizeBytes: size,
				}
			}
		}

		// 2. Check for default/legacy binary: e.g. engines/bin/mimir_tyr_engine.exe
		binPath := filepath.Join(rootDir, filepath.FromSlash(e.Binary))
		sum, size, err := FileSHA256(binPath)
		if err == nil {
			e.SHA256 = sum
			e.SizeBytes = size
			e.Version = probeVersion(ctx, binPath)
			// Also register in Platforms map under host platform if not already set
			hostKey := runtime.GOOS + "_" + runtime.GOARCH
			if _, exists := e.Platforms[hostKey]; !exists {
				e.Platforms[hostKey] = PlatformBinary{
					Binary:    e.Binary,
					SHA256:    sum,
					SizeBytes: size,
				}
			}
		}

		if len(e.Platforms) > 0 || err == nil {
			if e.Version == "" {
				e.Version = "0.2.0"
			}
			m.Engines = append(m.Engines, e)
		}
	}
	if len(m.Engines) == 0 {
		return m, errors.New("engines: no engine binaries found in engines/bin/")
	}
	raw, _ := json.MarshalIndent(m, "", "  ")
	out := filepath.Join(rootDir, "engines", "bin", "manifest.json")
	if err := os.WriteFile(out, append(raw, '\n'), 0o644); err != nil {
		return m, err
	}
	return m, nil
}

// EngineDiff describes one engine whose local copy differs from the published copy.
type EngineDiff struct {
	ID            string `json:"id"`
	LocalVersion  string `json:"local_version"`
	RemoteVersion string `json:"remote_version"`
	Reason        string `json:"reason"` // "new", "changed", "missing_local"
	Remote        EngineEntry
}

// UpdatePlan is the result of comparing local and remote manifests.
type UpdatePlan struct {
	Source  string       `json:"source"`
	Remote  Manifest     `json:"remote"`
	Changes []EngineDiff `json:"changes"`
}

func httpGet(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mimir-updater")
	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// CheckForUpdates fetches the published manifest and diffs it against local binaries
// by SHA-256 (not by version string), so any rebuilt engine is detected.
func CheckForUpdates(ctx context.Context, rootDir, base string) (UpdatePlan, error) {
	plan := UpdatePlan{Source: base}
	raw, err := httpGet(ctx, base+"/engines/bin/manifest.json", 1<<20)
	if err != nil {
		return plan, err
	}
	var remote Manifest
	if err := json.Unmarshal(raw, &remote); err != nil {
		return plan, fmt.Errorf("engines: remote manifest is not valid JSON: %w", err)
	}
	// Refuse to install anything from a manifest that drops Alpha Flux LLC ownership.
	if !strings.EqualFold(remote.Owner, RequiredOwner) || !strings.Contains(remote.Attribution, RequiredOwner) {
		return plan, fmt.Errorf("%w: remote manifest at %s", ErrAttributionRequired, base)
	}
	plan.Remote = remote

	local, _ := readManifestFile(rootDir)
	localByID := map[string]EngineEntry{}
	for _, e := range local.Engines {
		localByID[e.ID] = e
	}
	for _, re := range remote.Engines {
		le, known := localByID[re.ID]
		binPath := filepath.Join(rootDir, filepath.FromSlash(re.Binary))
		sum, _, ferr := FileSHA256(binPath)
		switch {
		case ferr != nil:
			plan.Changes = append(plan.Changes, EngineDiff{ID: re.ID, LocalVersion: "none", RemoteVersion: re.Version, Reason: "missing_local", Remote: re})
		case !known:
			if !strings.EqualFold(sum, re.SHA256) {
				plan.Changes = append(plan.Changes, EngineDiff{ID: re.ID, LocalVersion: "unknown", RemoteVersion: re.Version, Reason: "new", Remote: re})
			}
		case !strings.EqualFold(sum, re.SHA256):
			plan.Changes = append(plan.Changes, EngineDiff{ID: re.ID, LocalVersion: le.Version, RemoteVersion: re.Version, Reason: "changed", Remote: re})
		}
	}
	return plan, nil
}

func readManifestFile(rootDir string) (Manifest, error) {
	var m Manifest
	raw, err := os.ReadFile(filepath.Join(rootDir, "engines", "bin", "manifest.json"))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(raw, &m)
	return m, err
}

// ApplyUpdates downloads each changed binary to a temp file, verifies its SHA-256
// against the remote manifest, then atomically swaps it into place. Nothing is
// replaced unless every download verifies, and the manifest is written last.
func ApplyUpdates(ctx context.Context, rootDir string, plan UpdatePlan) ([]string, error) {
	if len(plan.Changes) == 0 {
		return nil, nil
	}
	type staged struct{ tmp, dst string }
	var ready []staged
	cleanup := func() {
		for _, s := range ready {
			_ = os.Remove(s.tmp)
		}
	}
	for _, c := range plan.Changes {
		data, err := httpGet(ctx, plan.Source+"/"+c.Remote.Binary, 256<<20)
		if err != nil {
			cleanup()
			return nil, err
		}
		dst := filepath.Join(rootDir, filepath.FromSlash(c.Remote.Binary))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			cleanup()
			return nil, err
		}
		tmp := dst + ".download"
		if err := os.WriteFile(tmp, data, 0o755); err != nil {
			cleanup()
			return nil, err
		}
		ready = append(ready, staged{tmp: tmp, dst: dst})
		sum, _, _ := FileSHA256(tmp)
		if !strings.EqualFold(sum, c.Remote.SHA256) {
			cleanup()
			return nil, fmt.Errorf("%w: downloaded %s sha256=%s want=%s", ErrBinaryIntegrity, c.Remote.Binary, sum, c.Remote.SHA256)
		}
	}
	var updated []string
	for i, s := range ready {
		_ = os.Remove(s.dst + ".old")
		if _, err := os.Stat(s.dst); err == nil {
			if err := os.Rename(s.dst, s.dst+".old"); err != nil {
				cleanup()
				return updated, err
			}
		}
		if err := os.Rename(s.tmp, s.dst); err != nil {
			return updated, err
		}
		_ = os.Remove(s.dst + ".old")
		updated = append(updated, plan.Changes[i].ID+" -> "+plan.Changes[i].RemoteVersion)
	}
	raw, _ := json.MarshalIndent(plan.Remote, "", "  ")
	if err := os.WriteFile(filepath.Join(rootDir, "engines", "bin", "manifest.json"), append(raw, '\n'), 0o644); err != nil {
		return updated, err
	}
	return updated, nil
}

// HarnessUpdateResult describes what happened to the open-source harness itself.
type HarnessUpdateResult struct {
	Mode    string // "git" or "skipped"
	Output  string
	Rebuilt bool
}

// UpdateHarness fast-forwards the local git checkout of main (no merges, no force)
// and rebuilds bin/mimir. On Windows the running exe is renamed aside first, which
// Windows allows, so the new build can take its place.
func UpdateHarness(ctx context.Context, rootDir string) (HarnessUpdateResult, error) {
	if _, err := os.Stat(filepath.Join(rootDir, ".git")); err != nil {
		return HarnessUpdateResult{Mode: "skipped", Output: "not a git checkout; download the latest release instead"}, nil
	}
	pull := exec.CommandContext(ctx, "git", "pull", "--ff-only", "origin", "main")
	pull.Dir = rootDir
	out, err := pull.CombinedOutput()
	res := HarnessUpdateResult{Mode: "git", Output: strings.TrimSpace(string(out))}
	if err != nil {
		return res, fmt.Errorf("git pull --ff-only failed (local changes or diverged history?): %w", err)
	}
	if strings.Contains(res.Output, "Already up to date") {
		return res, nil
	}
	if _, err := exec.LookPath("go"); err != nil {
		res.Output += "\n(go toolchain not found; rebuild bin/mimir manually)"
		return res, nil
	}
	exe := "mimir"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	target := filepath.Join(rootDir, "bin", exe)
	tmp := target + ".new"
	build := exec.CommandContext(ctx, "go", "build", "-o", tmp, "./cmd/mimir")
	build.Dir = rootDir
	if bout, berr := build.CombinedOutput(); berr != nil {
		return res, fmt.Errorf("rebuild failed: %v\n%s", berr, bout)
	}
	_ = os.Remove(target + ".old")
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, target+".old"); err != nil {
			return res, err
		}
	}
	if err := os.Rename(tmp, target); err != nil {
		return res, err
	}
	res.Rebuilt = true
	return res, nil
}
