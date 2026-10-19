package engines

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

var (
	// ErrInsecureUpdateBase is returned when the update base is not https
	// and the operator has not explicitly opted into an insecure channel.
	ErrInsecureUpdateBase = errors.New("engines: update base must use https (set MIMIR_UPDATE_ALLOW_INSECURE=1 to override)")
	// ErrDowngradeRefused is returned when an update would install an older
	// engine version than the one already present.
	ErrDowngradeRefused = errors.New("engines: refusing to downgrade an engine (set MIMIR_ENGINE_ALLOW_DOWNGRADE=1 to force)")
	// ErrInsecureRedirect is returned when an update download is redirected
	// to a non-https URL and the operator has not opted into an insecure
	// channel. Validating the initial URL is not enough: a redirect must
	// not be able to downgrade the transport after the check.
	ErrInsecureRedirect = errors.New("engines: update download redirected to a non-https URL (set MIMIR_UPDATE_ALLOW_INSECURE=1 to override)")
	// ErrTooManyRedirects is returned when an update download exceeds the
	// redirect cap.
	ErrTooManyRedirects = errors.New("engines: update download exceeded the maximum number of redirects")
	// ErrUnverifiableVersion is returned when an update targets an engine
	// that is already installed but either version string is unparseable,
	// so the updater cannot prove the update is not a downgrade.
	ErrUnverifiableVersion = errors.New("engines: cannot verify update direction for an installed engine with an unparseable version (set MIMIR_ENGINE_ALLOW_DOWNGRADE=1 to force)")
)

// maxUpdateRedirects caps how many redirects an update download follows.
const maxUpdateRedirects = 5

// updateRedirectPolicy is the redirect policy for every update download.
// A redirect target must stay on https unless the operator explicitly
// opted into an insecure channel, and the chain is capped so a hostile or
// broken host cannot bounce the client forever. req is the redirect
// target about to be fetched; via holds the requests already made, so
// len(via) is the number of the redirect being attempted.
func updateRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) > maxUpdateRedirects {
		return fmt.Errorf("%w: stopped after %d redirects", ErrTooManyRedirects, maxUpdateRedirects)
	}
	if req.URL.Scheme != "https" && os.Getenv("MIMIR_UPDATE_ALLOW_INSECURE") != "1" {
		return fmt.Errorf("%w: %s", ErrInsecureRedirect, req.URL.Redacted())
	}
	return nil
}

// updateHTTPClient builds the client used for all update downloads, with
// the redirect policy attached.
func updateHTTPClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Minute, CheckRedirect: updateRedirectPolicy}
}

// validateUpdateBase enforces transport security for the update channel
// (N7): the base serves executable code, so plain http is refused unless
// the operator explicitly sets MIMIR_UPDATE_ALLOW_INSECURE=1.
func validateUpdateBase(base string) error {
	u, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("engines: invalid update base %q: %w", base, err)
	}
	if u.Scheme == "https" {
		return nil
	}
	if os.Getenv("MIMIR_UPDATE_ALLOW_INSECURE") == "1" {
		fmt.Fprintf(os.Stderr, "WARNING: engine update base %q is not https. MIMIR_UPDATE_ALLOW_INSECURE=1 is set, so binaries will be downloaded over an insecure channel.\n", base)
		return nil
	}
	return fmt.Errorf("%w: got %q", ErrInsecureUpdateBase, base)
}

// ManifestSigningBytes returns the canonical bytes an ed25519 manifest
// signature covers: the manifest encoded with encoding/json (struct field
// order, no indentation) with the Signature field cleared. Verifiers
// re-encode the parsed manifest the same way, so the served file's own
// formatting does not matter.
//
// Release signing procedure for maintainers (N7):
//  1. Generate an ed25519 keypair once (crypto/ed25519.GenerateKey) and
//     keep the private key offline, outside the repo.
//  2. After BuildManifest writes engines/bin/manifest.json, load it, call
//     SignManifest(manifest, priv), store the result in the manifest's
//     "signature" field, and commit the signed manifest.
//  3. Give installs the public key out of band via the environment:
//     MIMIR_ENGINE_PUBKEY=<base64 raw 32-byte public key>. Clients with
//     the variable set refuse unsigned or badly signed manifests.
func ManifestSigningBytes(m Manifest) ([]byte, error) {
	m.Signature = ""
	return json.Marshal(m)
}

// SignManifest returns the base64 ed25519 signature for a manifest, in the
// form stored in Manifest.Signature.
func SignManifest(m Manifest, priv ed25519.PrivateKey) (string, error) {
	canonical, err := ManifestSigningBytes(m)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ed25519.Sign(priv, canonical)), nil
}

// engineUpdatePubKey reads the operator-pinned manifest verification key
// from MIMIR_ENGINE_PUBKEY (base64 raw public key). Nil means the operator
// has not pinned a key.
func engineUpdatePubKey() (ed25519.PublicKey, error) {
	v := strings.TrimSpace(os.Getenv("MIMIR_ENGINE_PUBKEY"))
	if v == "" {
		return nil, nil
	}
	raw, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		raw2, err2 := base64.RawStdEncoding.DecodeString(v)
		if err2 != nil {
			return nil, fmt.Errorf("engines: MIMIR_ENGINE_PUBKEY is not valid base64: %w", err)
		}
		raw = raw2
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("engines: MIMIR_ENGINE_PUBKEY must decode to %d bytes, got %d", ed25519.PublicKeySize, len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

// verifyRemoteManifest authenticates a fetched manifest (N7). With a
// pinned key the manifest must carry a valid signature. Without one, the
// update still works (hash-only integrity, as before) but a loud warning
// states plainly that it is unauthenticated.
func verifyRemoteManifest(m Manifest) error {
	pub, err := engineUpdatePubKey()
	if err != nil {
		return err
	}
	if pub == nil {
		fmt.Fprintln(os.Stderr, "WARNING: MIMIR_ENGINE_PUBKEY is not set, so this engine update is UNAUTHENTICATED. The manifest and binaries are trusted purely because they came from the update host. Set MIMIR_ENGINE_PUBKEY to require signed manifests.")
		return nil
	}
	if strings.TrimSpace(m.Signature) == "" {
		return errors.New("engines: MIMIR_ENGINE_PUBKEY is set but the remote manifest carries no signature; refusing update")
	}
	sig, err := base64.StdEncoding.DecodeString(m.Signature)
	if err != nil {
		return fmt.Errorf("engines: remote manifest signature is not valid base64: %w", err)
	}
	canonical, err := ManifestSigningBytes(m)
	if err != nil {
		return err
	}
	if !ed25519.Verify(pub, canonical, sig) {
		return errors.New("engines: remote manifest signature verification failed; refusing update")
	}
	return nil
}

// compareVersions orders dotted numeric versions ("0.3.0" vs "0.2.9").
// ok is false when either side is missing or not dotted numeric, in which
// case callers must not treat the pair as comparable.
func compareVersions(a, b string) (cmp int, ok bool) {
	parse := func(s string) ([]int, bool) {
		s = strings.TrimSpace(strings.TrimPrefix(s, "v"))
		if s == "" || s == "none" || s == "unknown" {
			return nil, false
		}
		parts := strings.Split(s, ".")
		out := make([]int, len(parts))
		for i, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil, false
			}
			out[i] = n
		}
		return out, true
	}
	av, aok := parse(a)
	bv, bok := parse(b)
	if !aok || !bok {
		return 0, false
	}
	for i := 0; i < len(av) || i < len(bv); i++ {
		var x, y int
		if i < len(av) {
			x = av[i]
		}
		if i < len(bv) {
			y = bv[i]
		}
		if x != y {
			if x < y {
				return -1, true
			}
			return 1, true
		}
	}
	return 0, true
}

// hasLocalVersion reports whether a plan change carries an actual local
// version record. CheckForUpdates marks engines that are not installed
// yet with LocalVersion "none" (no binary on disk) or "unknown" (a binary
// is present but the local manifest does not track the engine); neither
// is a local version.
func hasLocalVersion(v string) bool {
	return v != "" && v != "none" && v != "unknown"
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
	resp, err := updateHTTPClient().Do(req)
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
	if err := validateUpdateBase(base); err != nil {
		return plan, err
	}
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
	// Authenticate the manifest before trusting any hash in it (N7).
	if err := verifyRemoteManifest(remote); err != nil {
		return plan, err
	}
	plan.Remote = remote

	local, _ := readManifestFile(rootDir)
	localByID := map[string]EngineEntry{}
	for _, e := range local.Engines {
		localByID[e.ID] = e
	}
	for _, re := range remote.Engines {
		le, known := localByID[re.ID]
		// Diff the file Invoke will actually execute on this host (the
		// platform-resolved binary), not the legacy top-level path (N8).
		binRel, wantSHA, _, found := re.ResolveBinary()
		if !found || binRel == "" {
			continue // nothing published for this host platform
		}
		binPath := filepath.Join(rootDir, filepath.FromSlash(binRel))
		sum, _, ferr := FileSHA256(binPath)
		switch {
		case ferr != nil:
			plan.Changes = append(plan.Changes, EngineDiff{ID: re.ID, LocalVersion: "none", RemoteVersion: re.Version, Reason: "missing_local", Remote: re})
		case !known:
			if !strings.EqualFold(sum, wantSHA) {
				plan.Changes = append(plan.Changes, EngineDiff{ID: re.ID, LocalVersion: "unknown", RemoteVersion: re.Version, Reason: "new", Remote: re})
			}
		case !strings.EqualFold(sum, wantSHA):
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
	// Re-run the trust checks here too: a plan does not have to come from
	// CheckForUpdates, and nothing executable is fetched before these pass.
	if err := validateUpdateBase(plan.Source); err != nil {
		return nil, err
	}
	if err := verifyRemoteManifest(plan.Remote); err != nil {
		return nil, err
	}
	// Downgrade protection (N7): a manifest that would move an engine
	// backwards is refused unless the operator explicitly forces it.
	// The check fails closed for engines that are already installed
	// (audit FIX 2): when a local version is on record but either
	// version string is unparseable, no ordering can be proven, so the
	// update is refused rather than silently waved through. Engines
	// with no recorded local version (new installs and untracked
	// binaries) keep the permissive path.
	if os.Getenv("MIMIR_ENGINE_ALLOW_DOWNGRADE") != "1" {
		for _, c := range plan.Changes {
			cmp, ok := compareVersions(c.RemoteVersion, c.LocalVersion)
			switch {
			case ok && cmp < 0:
				return nil, fmt.Errorf("%w: %s %s -> %s", ErrDowngradeRefused, c.ID, c.LocalVersion, c.RemoteVersion)
			case !ok && hasLocalVersion(c.LocalVersion):
				return nil, fmt.Errorf("%w: %s local=%q remote=%q", ErrUnverifiableVersion, c.ID, c.LocalVersion, c.RemoteVersion)
			}
		}
	}
	type staged struct{ tmp, dst string }
	var ready []staged
	cleanup := func() {
		for _, s := range ready {
			_ = os.Remove(s.tmp)
		}
	}
	for _, c := range plan.Changes {
		// Download, verify, and replace the platform-resolved binary that
		// Invoke executes on this host, not the legacy top-level path (N8).
		binRel, wantSHA, _, found := c.Remote.ResolveBinary()
		if !found || binRel == "" {
			cleanup()
			return nil, fmt.Errorf("engines: update for %s has no binary for host platform %s_%s", c.ID, runtime.GOOS, runtime.GOARCH)
		}
		data, err := httpGet(ctx, plan.Source+"/"+binRel, 256<<20)
		if err != nil {
			cleanup()
			return nil, err
		}
		dst := filepath.Join(rootDir, filepath.FromSlash(binRel))
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
		if !strings.EqualFold(sum, wantSHA) {
			cleanup()
			return nil, fmt.Errorf("%w: downloaded %s sha256=%s want=%s", ErrBinaryIntegrity, binRel, sum, wantSHA)
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
